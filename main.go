package main

import (
	"fmt"
	"net/http"
)

type Router struct{}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	_ = r
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to my site!</h1>")
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

func (router Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		homeHandler(w, r)
	case "/contact":
		contactHandler(w, r)
	case "/faq":
		faqHandler(w, r)
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func main() {
	var router Router
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, router)
}
