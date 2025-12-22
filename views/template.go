package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type Template struct {
	htmlTpl *template.Template
}

func ParseFS(fs fs.FS, pattern string) (Template, error) {
	tpl, err := template.ParseFS(fs, pattern)
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

func Parse(filepath string) (Template, error) {
	tpl, err := template.ParseFiles(filepath)
	if err != nil {
		log.Printf("parsing template: %v", err)
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}
	return Template{htmlTpl: tpl}, nil
}

func (t Template) Execute(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	err := t.htmlTpl.Execute(w, nil)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
	}
}
