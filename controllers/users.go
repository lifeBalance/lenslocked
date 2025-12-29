package controllers

import (
	"fmt"
	"net/http"

	"github.com/lifebalance/lenslocked/context"
	"github.com/lifebalance/lenslocked/models"
)

type Users struct {
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService    *models.UserService
	SessionService *models.SessionService
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")   // parse query string
	u.Templates.New.Execute(w, r, data) // render email in the template
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		fmt.Println(err.Error()) // rudimentary logging
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Upsert(int(user.ID))
	if err != nil {
		fmt.Println(err.Error()) // rudimentary logging
		// TODO: show a warning about the issue
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieName, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err.Error()) // rudimentary logging
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	session, err := u.SessionService.Upsert(int(user.ID))
	if err != nil {
		fmt.Println(err.Error()) // rudimentary logging
		// TODO: show a warning about the issue
		return
	}
	setCookie(w, CookieName, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := context.User(ctx)
	fmt.Fprintf(w, "Current user: %s\n", user.Email)
	if user == nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
}

func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	sessionToken, err := readCookie(r, CookieName)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.DeleteSession(sessionToken)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	deleteCookie(w, CookieName)
	http.Redirect(w, r, "/signin", http.StatusFound)
}

type UserMiddleware struct {
	SessionService *models.SessionService
}

func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := readCookie(r, CookieName)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(sessionCookie)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
