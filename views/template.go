package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
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
			"csrfField": func() (template.HTML, error) {
				// return `<--! PLACEHOLDER: implement CSRF field in Execute-->`
				return "", fmt.Errorf("csrfField not implemented")
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

	// If we want to avoid half-rendered pages, render template to buffer,
	// and if there's no errors, flush the buffer on the ResponseWriter
	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
	}
	io.Copy(w, &buf)
}
