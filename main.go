package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/lifebalance/lenslocked/controllers"
	"github.com/lifebalance/lenslocked/views"
)

func renderTemplate(w http.ResponseWriter, filepath string) {
	t, err := views.Parse(filepath) // Call a package function
	if err != nil {
		log.Printf("parsing template: %v", err)
		http.Error(w, "Error parsing the template", http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil) // Call a views.Template method
}

// func homeHandler(w http.ResponseWriter, r *http.Request) {
// 	_ = r
// 	tplPath := filepath.Join("templates", "home.gohtml") // Windows
// 	renderTemplate(w, tplPath)
// }

// func contactHandler(w http.ResponseWriter, r *http.Request) {
// 	_ = r
// 	w.Header().Set("Content-type", "text/html; charset=utf-8")

// 	tplPath := filepath.Join("templates", "contact.gohtml") // Windows
// 	renderTemplate(w, tplPath)
// }

// func faqHandler(w http.ResponseWriter, r *http.Request) {
// 	_ = r
// 	w.Header().Set("Content-type", "text/html; charset=utf-8")
// 	tplPath := filepath.Join("templates", "faq.gohtml") // Windows
// 	renderTemplate(w, tplPath)
// }

func main() {
	r := chi.NewRouter()
	// r.Get("/", homeHandler)
	tpl, err := views.Parse(filepath.Join("templates", "home.gohtml"))
	if err != nil {
		panic(err)
	}
	r.Get("/", controllers.StaticHandler(tpl))

	// r.Get("/contact", contactHandler)
	tpl, err = views.Parse(filepath.Join("templates", "contact.gohtml"))
	if err != nil {
		panic(err)
	}
	r.Get("/contact", controllers.StaticHandler(tpl))

	// r.Get("/faq", faqHandler)
	tpl, err = views.Parse(filepath.Join("templates", "faq.gohtml"))
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.StaticHandler(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, r)
}
