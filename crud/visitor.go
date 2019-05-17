package crud

import (
	"bytes"
	"github.com/masseelch/gogen"
	"go/ast"
	"go/format"
	"os"
	"strings"
	"text/template"
	"time"
)

var (
	crudTpl = template.Must(template.New("").Funcs(gogen.FuncMap).Parse(string(kCrudTpl)))
)

type (
	TplData struct {
		Timestamp time.Time
		Type      string
	}

	Visitor struct {
		InputPath string

		// These are used to get information down the ast.
		GenDecl *ast.GenDecl
	}
)

func (v Visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch t := n.(type) {
	case *ast.GenDecl:
		v.GenDecl = t
	case *ast.TypeSpec:
		// Get the docs.
		var cg *ast.CommentGroup
		if t.Doc == nil {
			cg = v.GenDecl.Doc
		} else {
			cg = t.Doc
		}

		if strings.HasPrefix(cg.Text(), "gen:Handler") {
			if _, ok := t.Type.(*ast.StructType); ok {
				var buf bytes.Buffer
				d := TplData{
					Timestamp: time.Now(),
					Type:      t.Name.Name,
				}
				if err := crudTpl.Execute(&buf, d); err != nil {
					panic(err)
				}

				// Format generated code.
				g, err := format.Source(buf.Bytes())
				if err != nil {
					panic(err)
				}

				// Write interfaces to file.
				f, err := os.Create(gogen.MatchGoFile.ReplaceAllString(v.InputPath, "${1}.g.go"))
				if err != nil {
					panic(err)
				}
				defer f.Close()

				if _, err := f.Write(g); err != nil {
					panic(err)
				}
			}
		}
	}

	return v
}
