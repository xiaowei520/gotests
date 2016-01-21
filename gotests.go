package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/cweill/gotests/code"
	"github.com/cweill/gotests/output"
	"github.com/cweill/gotests/source"
)

var noTestsError = errors.New("no tests generated")

type funcs []string

func (f *funcs) String() string {
	return fmt.Sprint(*f)
}

func (f *funcs) Set(value string) error {
	if len(*f) > 0 {
		return errors.New("flag already set")
	}
	for _, fun := range strings.Split(value, ",") {
		*f = append(*f, fun)
	}
	return nil
}

var (
	onlyFlag, exclFlag funcs

	allFlag = flag.Bool("all", false, "generate tests for all functions in specified files or directories.")
)

func main() {
	flag.Var(&onlyFlag, "only", "comma-separated list of case-sensitive function names for which tests will be generating exclusively. Takes precedence over -all.")
	flag.Var(&exclFlag, "excl", "comma-separated list of case-sensitive function names to exclude when generating tests. Take precedence over -funcs and -all.")
	flag.Parse()
	if len(onlyFlag) == 0 && len(exclFlag) == 0 && !*allFlag {
		fmt.Println("Please specify either the -funcs or -all flag")
		return
	}
	if len(flag.Args()) == 0 {
		fmt.Println("Please specify a file or directory containing the source")
		return
	}
	for _, path := range flag.Args() {
		ps, err := source.Files(path)
		if err != nil {
			if err == source.NoFilesFound {
				fmt.Printf("No source files found at %v\n", path)
			} else {
				fmt.Println(err.Error())
			}
			continue
		}
		for _, src := range ps {
			tests, err := generateTests(string(src), src.TestPath(), onlyFlag, exclFlag)
			if err != nil {
				if err == noTestsError {
					fmt.Printf("No tests generated for %v\n", path)
				} else {
					fmt.Println(err.Error())
				}
				continue
			}
			for _, test := range tests {
				fmt.Printf("Generated %v\n", test)
			}
		}
	}
}

func generateTests(srcPath, destPath string, onlyFuncs, exclFuncs []string) ([]string, error) {
	info, err := code.Parse(srcPath)
	if err != nil {
		return nil, fmt.Errorf("code.Parse: %v", err)
	}
	funcs := info.TestableFuncs(onlyFuncs, exclFuncs)
	if len(funcs) == 0 {
		return nil, noTestsError
	}
	return output.Write(srcPath, destPath, info.Header, funcs)
}
