package crud

import "github.com/masseelch/gogen"

const kCrudTpl = gogen.KGeneratedFileWarningComment + `
package crud

import(
	"app"
	"app/server"
	"github.com/go-chi/chi"
	"github.com/masseelch/render"
	"net/http"
	"strconv"
)

func (h {{.Type}}) read(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		render.InternalServerError(w, err)
		return err
	}
	var e app.{{.Type}}
	if err := h.Repositories.{{.Type}}.Find(e, ) 
	render.JSON(w, "This is read {{.Type}}")
}
`
