package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
	"github.com/lifebalance/lenslocked/controllers"
	"github.com/lifebalance/lenslocked/migrations"
	"github.com/lifebalance/lenslocked/models"
	"github.com/lifebalance/lenslocked/templates"
	"github.com/lifebalance/lenslocked/views"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    []byte
		Secure bool
	}
	Server struct {
		Address string
	}
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Set up DB
	conn, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Migrations
	err = models.MigrateFS(conn, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// User services
	userService := &models.UserService{
		DB: conn,
	}
	sessionService := &models.SessionService{
		DB: conn,
	}
	passwordResetService := &models.PasswordResetService{
		DB: conn,
	}
	emailService, err := models.NewEmailService(cfg.SMTP)
	if err != nil {
		panic(err)
	}
	// Gallery services
	galleryService := &models.GalleryService{
		DB: conn,
	}

	// Set up the middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect(
		cfg.CSRF.Key,
		csrf.Path("/"),
		csrf.Secure(cfg.CSRF.Secure),
		csrf.TrustedOrigins([]string{
			"localhost:3000",
		}),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			log.Printf("CSRF validation failed!")
			log.Printf("Method: %s", r.Method)
			log.Printf("Path: %s", r.URL.Path)
			log.Printf("Origin: %s", r.Header.Get("Origin"))
			log.Printf("Referer: %s", r.Header.Get("Referer"))
			log.Printf("Token from form: %s", r.FormValue("gorilla.csrf.Token"))
			log.Println("failure reason:", csrf.FailureReason(r))

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

	// Users controllers
	usersController := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		PasswordResetService: passwordResetService,
		EmailService:         emailService,
	}
	usersController.Templates.New = views.MustParse(
		views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml"),
	)
	usersController.Templates.SignIn = views.MustParse(
		views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml"),
	)
	usersController.Templates.ForgotPassword = views.MustParse(
		views.ParseFS(templates.FS, "forgot-pwd.gohtml", "tailwind.gohtml"),
	)
	usersController.Templates.CheckYourEmail = views.MustParse(
		views.ParseFS(templates.FS, "check-your-email.gohtml", "tailwind.gohtml"),
	)
	usersController.Templates.ResetPassword = views.MustParse(
		views.ParseFS(templates.FS, "reset-pwd.gohtml", "tailwind.gohtml"),
	)
	// Galleries controllers
	galleriesController := controllers.Galleries{
		GalleryService: galleryService,
	}
	galleriesController.Templates.New = views.MustParse(
		views.ParseFS(
			templates.FS,
			"galleries/new.gohtml",
			"tailwind.gohtml",
		),
	)
	galleriesController.Templates.Edit = views.MustParse(
		views.ParseFS(
			templates.FS,
			"galleries/edit.gohtml",
			"tailwind.gohtml",
		),
	)

	// Set up router and routes
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)
	tpl := views.MustParse(views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml"))
	r.Get("/", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml"))
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl = views.MustParse(views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml"))
	r.Get("/faq", controllers.FAQ(tpl))

	r.Get("/signup", usersController.New)            // send the form
	r.Post("/users", usersController.Create)         // process the form
	r.Get("/signin", usersController.SignIn)         // send the form
	r.Post("/signin", usersController.ProcessSignIn) // process the form
	r.Post("/signout", usersController.ProcessSignOut)
	r.Get("/forgot-pwd", usersController.ForgotPassword)
	r.Post("/forgot-pwd", usersController.ProcessForgotPassword)
	r.Get("/reset-pwd", usersController.ResetPassword)
	r.Post("/reset-pwd", usersController.ProcessResetPassword)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersController.CurrentUser)
	})
	r.Route("/galleries", func(r chi.Router) {
		// Group is needed so that only CREATING galleries require an authenticated user
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/{id}/edit", galleriesController.Edit)    // send the form
			r.Post("/{id}/edit", galleriesController.Update) // process the form
			r.Get("/new", galleriesController.New)
			r.Post("/", galleriesController.Create)
		})
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on %s\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}

// Wrap any HandlerFunc with this mw to time it.
// func TimerMiddleware(h http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		startTime := time.Now()
// 		h(w, r)
// 		fmt.Println("Request time:", time.Since(startTime))
// 	}
// }

func loadSMTPConfig() (models.SMTPConfig, error) {
	portString := os.Getenv("MAILTRAP_PORT")
	portInt, err := strconv.Atoi(portString)
	if err != nil {
		portInt = 2525
	}
	cfg := models.SMTPConfig{
		Host: os.Getenv("MAILTRAP_HOST"),
		User: os.Getenv("MAILTRAP_USERNAME"),
		Pass: os.Getenv("MAILTRAP_PASSWORD"),
		Port: portInt,
	}
	if cfg.Host == "" || cfg.User == "" || cfg.Pass == "" {
		return cfg, fmt.Errorf("missing MAILTRAP_* envs")
	}
	return cfg, nil
}

func loadEnvConfig() (config, error) {
	var cfg config

	// Load env. variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return cfg, err
	}

	// PSQL
	cfg.PSQL = models.DefaultPostgresConfig() // TODO: Load from env

	// SMTP
	smtpConfig, err := loadSMTPConfig()
	if err != nil {
		log.Fatalf("failed to load SMTP config: %v", err)
		return cfg, fmt.Errorf("failed to load SMTP config: %v", err)
	}
	cfg.SMTP = smtpConfig

	// CSRF
	csrfSecureString := os.Getenv("CSRF_SECURE")
	csrfKeyString := os.Getenv("CSRF_KEY")
	if csrfKeyString == "" || csrfSecureString == "" {
		return cfg, fmt.Errorf("missing CSRF env. var.")
	}
	cfg.CSRF.Secure, err = strconv.ParseBool(csrfSecureString)
	if err != nil {
		log.Fatalf("woops %v", err)
	}
	cfg.CSRF.Key = []byte(csrfKeyString)

	// Server
	cfg.Server.Address = ":3000" //  TODO: Load from env

	return cfg, nil
}
