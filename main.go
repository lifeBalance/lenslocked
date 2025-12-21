package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func renderTemplate(w http.ResponseWriter, filepath string) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	tpl, err := template.ParseFiles(filepath)
	if err != nil {
		log.Printf("parsing template: %v", err)
		http.Error(w, "Error parsing the template", http.StatusInternalServerError)
		return
	}

	err = tpl.Execute(w, nil)
	if err != nil {
		log.Printf("executing template: %v", err)
		http.Error(w, "Error executing the template", http.StatusInternalServerError)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	tplPath := filepath.Join("templates", "home.gohtml") // Windows
	renderTemplate(w, tplPath)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	tplPath := filepath.Join("templates", "contact.gohtml") // Windows
	renderTemplate(w, tplPath)
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	tplPath := filepath.Join("templates", "faq.gohtml") // Windows
	renderTemplate(w, tplPath)
}

func main() {
	r := chi.NewRouter()
	r.Get("/", homeHandler)
	r.Get("/contact", contactHandler)
	r.Get("/faq", faqHandler)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, r)
}
