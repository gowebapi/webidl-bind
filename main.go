package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"wasm/generator/gowasm"
	"wasm/generator/types"

	"github.com/dennwc/webidl/ast"
	"github.com/dennwc/webidl/parser"
)

var args struct {
	output   string
	warnings bool
}

var stopErr = errors.New("stopping for previous error")

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
	conv := types.NewConvert()
	setup := &types.Setup{
		Package: "",
		Error:   failing,
		Warning: warning,
	}

	for _, name := range flag.Args() {
		if err := processFile(name, conv, setup); err != nil {
			return err
		}
	}
	if err := conv.EvaluateInput(); err != nil {
		return err
	}
	if err := conv.EvaluateOutput(); err != nil {
		return err
	}

	files, err := gowasm.WriteSource(conv)
	if err != nil {
		return err
	}

	for k, v := range files {
		fmt.Println("writing output file", k, "to", args.output)
		if err := ioutil.WriteFile(args.output, v, 0666); err != nil {
			return err
		}
	}

	return nil
}

func processFile(filename string, conv *types.Convert, setup *types.Setup) error {
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
		return stopErr
	}

	setup.Package = gowasm.FormatPkg(filename)
	if err := conv.Process(file, setup); err != nil {
		return err
	}
	return nil
}

func failing(base *ast.Base, format string, args ...interface{}) {
	dst := os.Stderr
	fmt.Fprint(dst, "error:")
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
	fmt.Fprint(dst, "warning:")
	if base != nil {
		fmt.Fprint(dst, base.Line, ":")
	}
	fmt.Fprintf(dst, format, values...)
	fmt.Fprint(dst, "\n")
}

func parseArgs() string {
	flag.BoolVar(&args.warnings, "log-warning", true, "log warnings")
	flag.StringVar(&args.output, "output", "", "output file")
	flag.Parse()
	if len(flag.Args()) == 0 {
		return "no input files on command line"
	}
	if args.output == "" {
		return "missing output file"
	}
	return ""
}
