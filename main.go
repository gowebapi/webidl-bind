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

	"github.com/gowebapi/webidlgenerator/gowasm"
	"github.com/gowebapi/webidlgenerator/transform"
	"github.com/gowebapi/webidlgenerator/types"

	"github.com/gowebapi/webidlparser/ast"
	"github.com/gowebapi/webidlparser/parser"
)

var args struct {
	outputPath string
	warnings   bool
	singlePkg  string
	insidePkg  string
	goBuild    string
}

var errStop = errors.New("too many errors")
var currentFilename string

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
	if err := conv.EvaluateInput(); err != nil {
		return err
	}
	if err := conv.EvaluateOutput(); err != nil {
		return err
	}
	if err := trans.Execute(conv); err != nil {
		return err
	}
	transform.RenameOverrideMethods(conv)

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
	return nil
}

func processFile(filename string, conv *types.Convert, setup *types.Setup) error {
	currentFilename = filename
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	file := parser.Parse(string(content))
	trouble := ast.GetAllErrorNodes(file)
	if len(trouble) > 0 {
		sort.SliceStable(trouble, func(i, j int) bool { return trouble[i].Line < trouble[j].Line })
		for _, e := range trouble {
			failing(e.NodeBase(), e.Message)
		}
		return errStop
	}

	setup.Package = gowasm.FormatPkg(filename, args.singlePkg)
	if err := conv.Process(file, setup); err != nil {
		return err
	}
	currentFilename = "<unknown file>"
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

func failing(base *ast.Base, format string, args ...interface{}) {
	dst := os.Stderr
	fmt.Fprint(dst, "error:", currentFilename, ":")
	if base != nil {
		fmt.Fprint(dst, base.Line, ":")
	}
	fmt.Fprintf(dst, format, args...)
	fmt.Fprint(dst, "\n")
}

func warning(base *ast.Base, format string, values ...interface{}) {
	if !args.warnings {
		return
	}
	dst := os.Stderr
	fmt.Fprint(dst, "warning:", currentFilename, ":")
	if base != nil {
		fmt.Fprint(dst, base.Line, ":")
	}
	fmt.Fprintf(dst, format, values...)
	fmt.Fprint(dst, "\n")
}

func parseArgs() string {
	flag.BoolVar(&args.warnings, "log-warning", true, "log warnings")
	flag.StringVar(&args.outputPath, "output", "", "output path")
	flag.StringVar(&args.insidePkg, "inside-package", "", "output path is inside current package")
	flag.StringVar(&args.singlePkg, "single-package", "", "all types to same package")
	flag.StringVar(&args.goBuild, "go-build", "", "execute go build in output folders")
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
	return ""
}

func pathExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
