package main

import (
	"errors"
	"flag"
	"github.com/masseelch/gogen/repo"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type flags struct {
	InputPath string
}

func main() {
	// Parse the flags.
	var fs flags
	flag.StringVar(&fs.InputPath, "in", "", "Path to a go-file or directory of go-files containing model-declarations.")

	flag.Parse()

	if fs.InputPath == "" {
		panic(errors.New("no input path given"))
	}

	genRepositoryCode(fs)
}

func genRepositoryCode(fs flags) {
	var fileInfos []os.FileInfo

	pathInfo, err := os.Stat(fs.InputPath)
	if err != nil {
		panic(err)
	}

	// If the path is a directory read all go-files.
	if pathInfo.IsDir() {
		fileInfos, err = ioutil.ReadDir(fs.InputPath)
		if err != nil {
			panic(err)
		}
	} else {
		fileInfos = []os.FileInfo{pathInfo}
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".go") && !strings.HasSuffix(fileInfo.Name(), ".g.go") {
			path := fs.InputPath
			if pathInfo.IsDir() {
				path = filepath.Join(path, fileInfo.Name())
			}

			// Parse the source file.
			parsedFile, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
			if err != nil {
				panic(err)
			}

			// Create a visitor for the current file.
			v := repo.Visitor{
				InputPath:         fs.InputPath,
				PathInfo:          pathInfo,
				FileInfo:          fileInfo,
			}

			ast.Walk(v, parsedFile)
		}
	}
}
