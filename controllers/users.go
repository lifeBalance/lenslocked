package controllers

import (
	"fmt"
	"net/http"
)

type Users struct {
	Templates struct {
		New Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email") // parse query string
	u.Templates.New.Execute(w, data)  // render email in the template
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	// err := r.ParseForm()
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusBadRequest)
	// 	return
	// }
	// fmt.Fprint(w, "Email:", r.PostForm.Get("email"))
	// fmt.Fprint(w, "\tpassword:", r.PostForm.Get("password"))
	fmt.Fprint(w, "Email:", r.FormValue("email")) // No need for ParseForm
	fmt.Fprint(w, "\npassword:", r.FormValue("password"))
}
