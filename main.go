package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lifebalance/lenslocked/controllers"
	"github.com/lifebalance/lenslocked/templates"
	"github.com/lifebalance/lenslocked/views"
)

func main() {
	r := chi.NewRouter()
	tpl := views.MustParse(views.ParseFS(templates.FS, "home.gohtml"))
	r.Get("/", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "contact.gohtml"))
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "faq.gohtml"))
	r.Get("/faq", controllers.FAQ(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, r)
}
