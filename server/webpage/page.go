package webpage

import (
	_ "embed"
	"html/template"
	"net/http"
	"path/filepath"
	txtemplate "text/template"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/ross96D/updater/share/utils"
)

//go:embed view_file.html
var viewFileHtml string

//go:embed ansi_up.mjs
var ansiUpJs string

//go:embed main.css
var mainCss string

func GetMIME(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".mjs", ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	default:
		return ""
	}
}

func WebHandlers(r chi.Router) {
	txtmpl := txtemplate.New("")
	tmpl := template.New("")

	_, err := tmpl.New("view_file.html").Parse(viewFileHtml)

	_, err = txtmpl.New("ansi_up.mjs").Parse(ansiUpJs)
	_, err = txtmpl.New("main.css").Parse(mainCss)

	utils.Assert(err == nil, "template.New.Parse failed %s", err)

	r.Get("/view/{file}", func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "file")
		if file == "" {
			w.WriteHeader(404)
			return
		}
		tmpl.ExecuteTemplate(w, "view_file.html", struct{ FileName string }{
			FileName: file,
		})
	})
	r.Get("/assets/{file}", func(w http.ResponseWriter, r *http.Request) {
		file := chi.URLParam(r, "file")
		t := txtmpl.Lookup(file)
		if t == nil {
			w.WriteHeader(404)
			return
		}
		mime := GetMIME(file)
		if mime != "" {
			w.Header().Add("Content-Type", mime)
		}
		t.Execute(w, nil)
	})
	r.Handle(
		"/ws/{file}",
		WebSocketHandler{
			Upgrader: websocket.Upgrader{
				ReadBufferSize:  4086,
				WriteBufferSize: 4086,
			},
		},
	)

}
