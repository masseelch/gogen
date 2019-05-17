package repo

import (
	"bytes"
	"github.com/masseelch/gogen"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	intTpl      = template.Must(template.New("").Funcs(gogen.FuncMap).Parse(string(kInterfaceTpl)))
	implTpl     = template.Must(template.New("").Funcs(gogen.FuncMap).Parse(string(kImplementationTpl)))
)

type (
	InterfaceTplData struct {
		Timestamp time.Time
		Type      string
	}

	ImplementationTplData struct {
		Timestamp          time.Time
		Type               string
		InsertFields       []string
		InsertValues       []interface{}
		SelectDestinations []interface{}
		SelectFields       []string
		UpdateFields       []string
		UpdateValues       []interface{}
	}

	Visitor struct {
		InputPath string
		PathInfo  os.FileInfo
		FileInfo  os.FileInfo

		repositories []string

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

		if strings.HasPrefix(cg.Text(), "gen:Model") {
			if st, ok := t.Type.(*ast.StructType); ok {
				filename := v.InputPath
				if v.PathInfo.IsDir() {
					filename = filepath.Join(filename, v.FileInfo.Name())
				}

				v.repositories = append(v.repositories, t.Name.Name)

				// Generate the interfaces.
				v.generateInterfaces(t, filename)

				// Generate the implementations.
				v.generateImplementations(t, st, filename)
			}
		}
	}

	return v
}

func (v Visitor) generateInterfaces(t *ast.TypeSpec, filename string) {
	var buf bytes.Buffer
	d := InterfaceTplData{
		Timestamp: time.Now(),
		Type:      t.Name.Name,
	}
	if err := intTpl.Execute(&buf, d); err != nil {
		panic(err)
	}

	// Format generated code.
	g, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Write interfaces to file.
	f, err := os.Create(gogen.MatchGoFile.ReplaceAllString(filename, "${1}.g.go"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(g); err != nil {
		panic(err)
	}
}

func (v Visitor) generateImplementations(t *ast.TypeSpec, st *ast.StructType, filename string) {
	var insertFields []string
	var insertValues []interface{}
	var selectDestinations []interface{}
	var selectFields []string
	var updateFields []string
	var updateValues []interface{}
	for _, field := range st.Fields.List {
		if c := field.Comment.Text(); strings.HasPrefix(c, "gen:") {
			if strings.Contains(c, "select") {
				selectFields = append(selectFields, gogen.ToSnakeCase(field.Names[0].Name))
				selectDestinations = append(selectDestinations, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				insertFields = append(insertFields, gogen.ToSnakeCase(field.Names[0].Name))
				insertValues = append(insertValues, field.Names[0].Name)
			}
			if strings.Contains(c, "insert") {
				updateFields = append(updateFields, gogen.ToSnakeCase(field.Names[0].Name))
				updateValues = append(updateValues, field.Names[0].Name)
			}
		}
	}

	// Generate the code.
	var buf bytes.Buffer
	d := ImplementationTplData{
		Timestamp:          time.Now(),
		Type:               t.Name.Name,
		InsertFields:       insertFields,
		InsertValues:       insertValues,
		SelectDestinations: selectDestinations,
		SelectFields:       selectFields,
		UpdateFields:       updateFields,
		UpdateValues:       updateValues,
	}
	if err := implTpl.Execute(&buf, d); err != nil {
		panic(err)
	}

	// Format generated code.
	g, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Write interfaces to file.
	f, err := os.Create(strings.ReplaceAll(
		filename,
		gogen.MatchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}"),
		"sql/"+gogen.MatchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}.g"),
	))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if _, err := f.Write(g); err != nil {
		panic(err)
	}
}
