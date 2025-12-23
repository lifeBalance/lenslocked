package controllers

import (
	"net/http"

	"github.com/lifebalance/lenslocked/views"
)

type Users struct {
	Templates struct {
		New views.Template
	}
}

func (u Users) New(w http.ResponseWriter, r *http.Request) {
	// Render the view with the signup form
	u.Templates.New.Execute(w, nil)
}
