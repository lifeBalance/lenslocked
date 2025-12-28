package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/lifebalance/lenslocked/controllers"
	"github.com/lifebalance/lenslocked/models"
	"github.com/lifebalance/lenslocked/rand"
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
	sessionService := models.SessionService{
		DB: conn,
	}
	usersController := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	// SIGN UP
	usersController.Templates.New = views.MustParse(views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"))
	r.Get("/signup", usersController.New) // send the form (/users/new is an alternative)
	r.Post("/users", usersController.Create)

	// SIGN IN
	usersController.Templates.SignIn = views.MustParse(views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"))
	r.Get("/signin", usersController.SignIn) // send the form (/sessions/new is an alternative)
	r.Post("/signin", usersController.ProcessSignIn)

	// SIGN OUT
	r.Post("/signout", usersController.ProcessSignOut)

	// Cookies
	r.Get("/users/me", usersController.CurrentUser)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})
	const PORT string = ":3000"

	fmt.Println("Starting the server on", PORT)
	// SETUP CRSF
	csrfKey, err := rand.RandomBytes(32)
	if err != nil {
		log.Fatalf("failed to generate CSRF key: %v", err)
	}

	csrfMw := csrf.Protect(
		csrfKey,
		csrf.Secure(false), // fix this before deploying
		csrf.TrustedOrigins([]string{
			"localhost:3000",
			"localhost:3000/signup",
			"localhost:3000/signin",
		}),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("CSRF validation failed!")
			log.Printf("Method: %s", r.Method)
			log.Printf("Path: %s", r.URL.Path)
			log.Printf("Origin: %s", r.Header.Get("Origin"))
			log.Printf("Referer: %s", r.Header.Get("Referer"))
			log.Printf("Token from form: %s", r.FormValue("gorilla.csrf.Token"))

			// Check for CSRF cookie
			cookie, err := r.Cookie("_gorilla_csrf")
			if err != nil {
				log.Printf("CSRF cookie error: %v", err)
			} else {
				log.Printf("CSRF cookie value: %s", cookie.Value)
			}

			http.Error(w, "CSRF token invalid", http.StatusForbidden)
		})),
	)
	http.ListenAndServe(PORT, csrfMw(r))
}

// Wrap any HandlerFunc with this mw to time it.
// func TimerMiddleware(h http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		startTime := time.Now()
// 		h(w, r)
// 		fmt.Println("Request time:", time.Since(startTime))
// 	}
// }
