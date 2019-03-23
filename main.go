package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gowebapi/webidl-bind/gowasm"
	"github.com/gowebapi/webidl-bind/transform"
	"github.com/gowebapi/webidl-bind/types"
)

var args struct {
	outputPath string
	warnings   bool
	singlePkg  string
	insidePkg  string
	goBuild    string
	goTest     string
	statusFile string
}

var errStop = errors.New("too many errors")

func main() {
	if msg := parseArgs(); msg != "" {
		fmt.Fprintln(os.Stderr, "command line error:", msg)
		os.Exit(1)
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if fi, err := os.Stat(args.outputPath); err != nil {
		return fmt.Errorf("trouble evaluate %s: %s", args.outputPath, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("output path '%s' doesn't point to a directory", args.outputPath)
	}

	trans := transform.New()
	conv := types.NewConvert()
	setup := &types.Setup{
		Package: args.singlePkg,
		Error:   failing,
		Warning: warning,
	}

	for _, name := range flag.Args() {
		ext := filepath.Ext(name)
		if ext == ".md" {
			fmt.Println("reading modificaton file", name)
			pkg := gowasm.FormatPkg(name, args.singlePkg)
			if err := trans.Load(name, pkg); err != nil {
				return err
			}
		} else {
			fmt.Println("reading WebIDL file", name)
			if err := processFile(name, conv, setup); err != nil {
				return err
			}
		}
	}
	if err := conv.Evaluate(); err != nil {
		return err
	}
	if err := trans.Execute(conv); err != nil {
		return err
	}
	transform.RenameOverrideMethods(conv)
	conv.Sort()

	files, err := gowasm.WriteSource(conv)
	if err != nil {
		return err
	}

	folders := []string{}
	for _, src := range files {
		filename, inc := src.Filename(args.insidePkg)
		if !inc {
			fmt.Printf("skipping '%s' as we are inside '%s'\n", src.Package, args.insidePkg)
			continue
		}
		path := filepath.Join(args.outputPath, filename)
		dir := filepath.Dir(path)
		folders = append(folders, dir)
		if !pathExist(dir) {
			fmt.Println("creating folder", dir)
			if err := os.MkdirAll(dir, 0775); err != nil {
				return err
			}
		}
		fmt.Println("saving ", path)
		if err := ioutil.WriteFile(path, src.Content, 0666); err != nil {
			return err
		}
	}
	if err := tryCompileResult(folders); err != nil {
		return err
	}
	if err := tryTestResult(folders); err != nil {
		return err
	}
	if args.statusFile != "" {
		if err := trans.WriteMarkdownStatus(args.statusFile); err != nil {
			return err
		}
	}
	return nil
}

func processFile(filename string, conv *types.Convert, setup *types.Setup) error {
	setup.Package = gowasm.FormatPkg(filename, args.singlePkg)
	setup.Filename = filename
	if err := conv.Load(setup); err != nil {
		return err
	}
	return nil
}

func tryCompileResult(folders []string) error {
	if args.goBuild == "" {
		return nil
	}
	sort.Strings(folders)
	last := ":/:"
	failed := []string{}
	for _, folder := range folders {
		if folder == last {
			continue
		}
		last = folder

		wasm := args.goBuild == "wasm"
		args := []string{"build"}
		if !wasm {
			args = append(args, "-i")
		}

		p := exec.Command("go", args...)
		p.Dir = folder
		p.Stdout = os.Stdout
		p.Stderr = os.Stderr
		if wasm {
			p.Env = os.Environ()
			p.Env = append(p.Env, "GOOS=js")
			p.Env = append(p.Env, "GOARCH=wasm")
		}
		fmt.Printf("> running '%s' in folder %s\n", strings.Join(p.Args, " "), folder)
		if err := p.Run(); err != nil {
			fmt.Println("> error: command failed:", err)
			failed = append(failed, folder)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("not all building was successful. failure in %s", strings.Join(failed, ", "))
	}
	return nil
}

func tryTestResult(folders []string) error {
	if args.goTest == "" {
		return nil
	}
	sort.Strings(folders)
	last := ":/:"
	failed := []string{}
	for _, folder := range folders {
		if folder == last {
			continue
		}
		last = folder

		// any test files?
		if yes, err := haveTestFiles(folder); err != nil {
			return err
		} else if !yes {
			continue
		}

		wasm := args.goTest == "wasm"
		args := []string{"test"}
		// if !wasm {
		// 	args = append(args, "-i")
		// }

		p := exec.Command("go", args...)
		p.Dir = folder
		p.Stdout = os.Stdout
		p.Stderr = os.Stderr
		if wasm {
			p.Env = os.Environ()
			p.Env = append(p.Env, "GOOS=js")
			p.Env = append(p.Env, "GOARCH=wasm")
		}
		fmt.Printf("> running '%s' in folder %s\n", strings.Join(p.Args, " "), folder)
		if err := p.Run(); err != nil {
			fmt.Println("> error: command failed:", err)
			failed = append(failed, folder)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("not all test was successful. failure in %s", strings.Join(failed, ", "))
	}
	return nil
}

func failing(ref types.GetRef, format string, args ...interface{}) {
	source := ""
	if ref != nil {
		where := ref.SourceReference()
		source = where.String() + ":"
	}
	dst := os.Stderr
	fmt.Fprint(dst, "error:", source)
	fmt.Fprintf(dst, format, args...)
	fmt.Fprint(dst, "\n")
}

func warning(ref types.GetRef, format string, values ...interface{}) {
	if !args.warnings {
		return
	}
	source := ""
	if ref != nil {
		where := ref.SourceReference()
		source = where.String() + ":"
	}
	dst := os.Stderr
	fmt.Fprint(dst, "warning:", source)
	fmt.Fprintf(dst, format, values...)
	fmt.Fprint(dst, "\n")
}

func parseArgs() string {
	flag.BoolVar(&args.warnings, "log-warning", true, "log warnings")
	flag.StringVar(&args.outputPath, "output", "", "output path")
	flag.StringVar(&args.insidePkg, "inside-package", "", "output path is inside current package")
	flag.StringVar(&args.singlePkg, "single-package", "", "all types to same package")
	flag.StringVar(&args.goBuild, "go-build", "", "execute go build in output folders")
	flag.StringVar(&args.goTest, "go-test", "", "execute go test in output folders")
	flag.StringVar(&args.statusFile, "spec-status", "", "write a markdown spec status file")
	flag.Parse()
	if len(flag.Args()) == 0 {
		return "no input files on command line"
	}
	if args.outputPath == "" {
		return "missing output path for file(s)"
	}
	if args.goBuild != "" && args.goBuild != "wasm" && args.goBuild != "host" {
		return "-go-build value should be 'wasm' or 'host'"
	}
	if args.goTest != "" && args.goTest != "wasm" && args.goTest != "host" {
		return "-go-test value should be 'wasm' or 'host'"
	}
	return ""
}

func pathExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func haveTestFiles(path string) (bool, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false, err
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_test.go") {
			return true, nil
		}
	}
	return false, nil
}
