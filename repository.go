package gogenrepo

import (
	"bytes"
	"go/ast"
	"go/format"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var (
	matchGoFile = regexp.MustCompile("(.)\\.go$")
	intTpl      = template.Must(template.New("").Parse(string(kInterfaceTpl)))
	implTpl     = template.Must(template.New("").Parse(string(kImplementationTpl)))
)

type RepositoryInterfaceTplData struct {
	Timestamp time.Time
	Package   string
	Type      string
}

type RepositoryImplementationTplData struct {
	Timestamp         time.Time
	ModelPackage      string
	Type              string
	InsertFields      []string
	InsertValues      []interface{}
	SelectDestination []interface{}
	SelectFields      []string
	UpdateFields      []string
	UpdateValues      []interface{}
}

type RepositoryVisitor struct {
	InputPath string
	PathInfo  os.FileInfo
	FileInfo  os.FileInfo

	// These are used to get information down the ast.
	Package string
	GenDecl *ast.GenDecl
}

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
					filename += "/" + v.FileInfo.Name()
				}

				// Generate the code.
				var intCode bytes.Buffer
				intData := RepositoryInterfaceTplData{
					Timestamp: time.Now(),
					Type:      t.Name.Name,
					Package:   v.Package,
				}
				if err := intTpl.Execute(&intCode, intData); err != nil {
					panic(err)
				}

				// Format generated code.
				g, err := format.Source(intCode.Bytes())
				if err != nil {
					panic(err)
				}

				// Write interfaces to file.
				intFile, err := os.Create(matchGoFile.ReplaceAllString(filename, "${1}.g.go"))
				if err != nil {
					panic(err)
				}

				if _, err := intFile.Write(g); err != nil {
					intFile.Close()
					panic(err)
				}
				intFile.Close()

				// Get the fields.
				var insertFields []string
				var insertValues []interface{}
				var selectDestinations []interface{}
				var selectFields []string
				var updateFields []string
				var updateValues []interface{}
				for _, field := range st.Fields.List {
					if c := field.Comment.Text(); strings.HasPrefix(c, "gen:") {
						if strings.Contains(c, "select") {
							selectFields = append(selectFields, toSnakeCase(field.Names[0].Name))
							selectDestinations = append(selectDestinations, field.Names[0].Name)
						}
						if strings.Contains(c, "insert") {
							insertFields = append(insertFields, toSnakeCase(field.Names[0].Name))
							insertValues = append(insertValues, field.Names[0].Name)
						}
						if strings.Contains(c, "insert") {
							updateFields = append(updateFields, toSnakeCase(field.Names[0].Name))
							updateValues = append(updateValues, field.Names[0].Name)
						}
					}
				}

				// Generate the code.
				implData := RepositoryImplementationTplData{
					Timestamp:         time.Now(),
					ModelPackage:      v.Package,
					Type:              t.Name.Name,
					InsertFields:      insertFields,
					InsertValues:      insertValues,
					SelectDestination: selectDestinations,
					SelectFields:      selectFields,
					UpdateFields:      updateFields,
					UpdateValues:      updateValues,
				}
				var implCode bytes.Buffer
				if err := implTpl.Execute(&implCode, implData); err != nil {
					panic(err)
				}

				// Format generated code.
				g, err = format.Source(implCode.Bytes())
				if err != nil {
					panic(err)
				}

				// Write interfaces to file.
				matchGoFile.ReplaceAllString(filename, "${1}")
				implFile, err := os.Create(strings.ReplaceAll(
					filename,
					matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}"),
					"sql/"+matchGoFile.ReplaceAllString(v.FileInfo.Name(), "${1}.g"),
				))
				if err != nil {
					panic(err)
				}

				if _, err := implFile.Write(g); err != nil {
					implFile.Close()
					panic(err)
				}
				implFile.Close()
			}
		}
	}

	return v
}
