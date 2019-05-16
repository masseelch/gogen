package repo

import (
	"bytes"
	"github.com/masseelch/gogen"
	"go/ast"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	matchGoFile = regexp.MustCompile("(.)\\.go$")
	intTpl      = template.Must(template.New("").Parse(string(kInterfaceTpl)))
	implTpl     = template.Must(template.New("").Parse(string(kImplementationTpl)))
	crudTpl     = template.Must(template.New("").Parse(string(kCrudTpl)))
)

type (
	CrudTplData struct {
		Timestamp    time.Time
		Type         string
	}

	RepositoryInterfaceTplData struct {
		Timestamp time.Time
		Type      string
	}

	RepositoryImplementationTplData struct {
		Timestamp          time.Time
		Type               string
		InsertFields       []string
		InsertValues       []interface{}
		SelectDestinations []interface{}
		SelectFields       []string
		UpdateFields       []string
		UpdateValues       []interface{}
	}

	RepositoryVisitor struct {
		InputPath string
		PathInfo  os.FileInfo
		FileInfo  os.FileInfo

		// These are used to get information down the ast.
		GenDecl *ast.GenDecl
	}
)

func (v RepositoryVisitor) Visit(n ast.Node) ast.Visitor {
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

				// Generate the interfaces.
				v.generateInterfaces(t, filename)

				// Generate the implementations.
				v.generateImplementations(t, st, filename)

				// Generate the crud handlers.
				v.generateCrudHandlers(t, filename)
			}
		}
	}

	return v
}

func (v RepositoryVisitor) generateCrudHandlers(t *ast.TypeSpec, filename string) {
	var buf bytes.Buffer
	d := CrudTplData{
		Timestamp:    time.Now(),
		Type:         t.Name.Name,
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
	f, err := os.Create(strings.ReplaceAll(
		filename,
		matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}"),
		"handler/"+matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}.g"),
	))
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	if _, err := f.Write(g); err != nil {
		f.Close()
		panic(err)
	}
	f.Close()
}

func (v RepositoryVisitor) generateInterfaces(t *ast.TypeSpec, filename string) {
	var buf bytes.Buffer
	d := RepositoryInterfaceTplData{
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
	f, err := os.Create(matchGoFile.ReplaceAllString(filename, "${1}.g.go"))
	if err != nil {
		panic(err)
	}

	if _, err := f.Write(g); err != nil {
		f.Close()
		panic(err)
	}
	f.Close()
}

func (v RepositoryVisitor) generateImplementations(t *ast.TypeSpec, st *ast.StructType, filename string) {
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
	d := RepositoryImplementationTplData{
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
		matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}"),
		"sql/"+matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}.g"),
	))
	if err != nil {
		panic(err)
	}

	if _, err := f.Write(g); err != nil {
		f.Close()
		panic(err)
	}
	f.Close()
}
