package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/gorilla/csrf"
	"github.com/lifebalance/lenslocked/context"
	"github.com/lifebalance/lenslocked/models"
)

type publicError interface {
	Public() string
}

type Template struct {
	htmlTpl *template.Template
}

func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	// Debugging templates in nested folders (galleries/new.gohtml)
	// fmt.Println("1", patterns)    // 1 [galleries/new.gohtml tailwind.gohtml]
	// fmt.Println("2", patterns[0]) // 2 galleries/new.gohtml

	// 'Base' extracts the filename, w/o the parent folder.
	// fmt.Println("3", path.Base(patterns[0])) // 3 new.gohtml

	tpl := template.New(path.Base(patterns[0]))
	tpl = tpl.Funcs(
		template.FuncMap{
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("csrfField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("currentUser not implemented")
			},
			"errors": func() []string {
				return nil
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

func (t Template) Execute(
	w http.ResponseWriter,
	r *http.Request,
	data interface{},
	errs ...error,
) {
	tpl, err := t.htmlTpl.Clone()
	if err != nil {
		log.Printf("cloning template: %v", err)
		http.Error(w, "Error cloning the template", http.StatusInternalServerError)
	}
	errMsgs := extractErrorMessages(errs...)
	tpl = tpl.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrf.TemplateField(r)
		},
		"currentUser": func() *models.User {
			return context.User(r.Context())
		},
		"errors": func() []string {
			return errMsgs
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

func extractErrorMessages(errs ...error) []string {
	var messages []string
	for _, err := range errs {
		var pubErr publicError
		if errors.As(err, &pubErr) {
			messages = append(messages, pubErr.Public()) // Extract msgs
		} else {
			fmt.Println(err) // rudimentary logging
			messages = append(messages, "Something went wrong")
		}
	}
	return messages
}
