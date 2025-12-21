package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	tplPath := filepath.Join("templates", "home.gohtml") // Windows
	tpl, err := template.ParseFiles(tplPath)
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

func contactHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Contact Page</h1><p>To get in touch, email me at <a href=\"mailto:sponge@bob.io\">sponge@bob.io</a></p>")
}

func faqHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<h1>FAQ PAGE</h1>
<ul>
<li>
	<p><strong>Q:</strong>Is there a free version?</p>
	<p><strong>A:</strong>Yes! We offer a free trial for 30 days on any paid plans</p>
</li>
<li>
	<p><strong>Q:</strong>What are your suppport hours?</p>
	<p><strong>A:</strong>We have support staff answering emails 24/7, though response times may be a bit slower on weekends.</p>
</li>
<li>
	<p><strong>Q:</strong>How do I contact suppport?</p>
	<p><strong>A:</strong>Email us - <a href="mailto:support@lenslocked.com">support@lenslocked.com</a></p>
</li>
</ul>
	`)
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
