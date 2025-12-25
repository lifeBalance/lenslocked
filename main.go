package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lifebalance/lenslocked/controllers"
	"github.com/lifebalance/lenslocked/models"
	"github.com/lifebalance/lenslocked/templates"
	"github.com/lifebalance/lenslocked/views"
)

func main() {
	r := chi.NewRouter()
	tpl := views.MustParse(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))
	r.Get("/", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))
	r.Get("/faq", controllers.FAQ(tpl))

	// Connecting to db
	cfg := models.DefaultPostgresConfig()
	conn, err := sql.Open("pgx", cfg.Stringify())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Initializing users service with the DB connection
	userService := models.UserService{
		DB: conn,
	}
	usersController := controllers.Users{
		UserService: &userService,
	}
	// SIGN UP
	usersController.Templates.New = views.MustParse(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	r.Get("/signup", usersController.New) // send the form (/users/new is an alternative)
	r.Post("/users", usersController.Create)

	// SIGN IN
	usersController.Templates.SignIn = views.MustParse(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	r.Get("/signin", usersController.SignIn) // send the form (/sessions/new is an alternative)
	r.Post("/signin", usersController.ProcessSignIn)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	http.ListenAndServe(PORT, r)
}
