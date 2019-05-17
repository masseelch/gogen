package main

import (
	"errors"
	"flag"
	"github.com/masseelch/gogen/crud"
	"go/ast"
	"go/parser"
	"go/token"
)

type flags struct {
	InputPath string
}

func main() {
	// Parse the flags.
	var fs flags
	flag.StringVar(&fs.InputPath, "in", "", "Path to a go-file containing a handler-struct.")

	flag.Parse()

	if fs.InputPath == "" {
		panic(errors.New("no input file given"))
	}

	// Parse the source-file
	src, err := parser.ParseFile(token.NewFileSet(), fs.InputPath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Create a visitor for the current file.
	v := crud.Visitor{
		InputPath: fs.InputPath,
	}

	ast.Walk(v, src)
}
