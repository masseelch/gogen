package repo

import "github.com/masseelch/gogen"

const kInterfaceTpl = gogen.KGeneratedFileWarningComment + `
package app

import (
	"github.com/Masterminds/squirrel"
)

type {{.Type}}Repository interface {
	CreateSelectBuilder(alias string) squirrel.SelectBuilder
	Find(*{{.Type}}, uint) error
    GetInsertFields() []string
    GetInsertValues({{.Type}}) []interface{}
    GetSelectDestinations({{.Type}}) []interface{}
    GetSelectFields(alias string) []string
    GetUpdateFields() []string
    GetUpdateValues({{.Type}}) []interface{}
}
`

const kImplementationTpl = gogen.KGeneratedFileWarningComment + `
package sql

import( 
	"app"
	"database/sql"
	"github.com/Masterminds/squirrel"
)

type {{.Type}} struct {
	DB sql.DB
}

func (r {{.Type}}) CreateSelectBuilder(alias string) squirrel.SelectBuilder {
	return squirrel.Select(r.GetSelectFields()...).From("{{.Type | ToLower }}_table")
}

func (r {{.Type}}) Find(e *app.{{.Type}}, id uint) error {
	return r.CreateSelectBuilder("e").Where("e.id", id).RunWith(r.DB).QueryRow().Scan(r.GetSelectDestinations(e))
}

func (r {{.Type}}) GetInsertFields(alias string) []string {
    return []string{ {{range .InsertFields}}alias+".{{.}}",{{end}} }
}

func (r {{.Type}}) GetInsertValues(m app.{{.Type}}) []interface{} {
    return []interface{}{ {{range .InsertValues}}m.{{.}},{{end}} }
}

func (r {{.Type}}) GetSelectDestinations(m *app.{{.Type}}) []interface{} {
	return []interface{}{ {{range .SelectDestinations}}&m.{{.}},{{end}} }
}

func (r {{.Type}}) GetSelectFields(alias string) []string {
    return []string{ {{range .SelectFields}}alias+".{{.}}",{{end}} }
}

func (r {{.Type}}) GetUpdateFields(alias string) []string {
    return []string{ {{range .UpdateFields}}alias+".{{.}}",{{end}} }
}

func (r {{.Type}}) GetUpdateValues(m app.{{.Type}}) []interface{} {
    return []interface{}{ {{range .UpdateValues}}m.{{.}},{{end}} }
}
`
