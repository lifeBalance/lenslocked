package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

type Template struct {
	htmlTpl *template.Template
}

func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	tpl := template.New(patterns[0])
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return `<--! PLACEHOLDER: implement CSRF field in Execute-->`
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		log.Printf("parsing FS template: %v", err)
		return Template{}, fmt.Errorf("parsing FS template: %w", err)
	}
	return Template{htmlTpl: tpl}, nil
}

func MustParse(t Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return t
}

// func Parse(filepath string) (Template, error) {
// 	tpl, err := template.ParseFiles(filepath)
// 	if err != nil {
// 		log.Printf("parsing template: %v", err)
// 		return Template{}, fmt.Errorf("parsing template: %w", err)
// 	}
// 	return Template{htmlTpl: tpl}, nil
// }

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "Error cloning the template", http.StatusInternalServerError)
	}
	tpl = tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
	})
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
	}
}
