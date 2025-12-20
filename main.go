package main

import (
	"fmt"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to my site!</h1>")
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Contact Page</h1><p>To get in touch, email me at <a href=\"mailto:sponge@bob.io\">sponge@bob.io</a></p>")
}

func pathHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	default:
		// w.WriteHeader(http.StatusNotFound)
		// fmt.Fprint(w, "Page not found")
		// http.Error(w, "Page not found", http.StatusNotFound) // Shorter
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound) // Standard not found message
	}
}

func main() {
	const PORT string = ":3000"
	http.HandleFunc("/", pathHandler)
	// http.HandleFunc("/contact", contactHandler)
	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, nil)
}
