package controllers

import (
	"fmt"
	"net/http"

	"github.com/lifebalance/lenslocked/context"
	"github.com/lifebalance/lenslocked/models"
)

type Galleries struct {
	Templates struct {
		New Template
	}
	GalleryService *models.GalleryService
}

// Render form to create a new gallery
func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")   // parse query string
	g.Templates.New.Execute(w, r, data) // render title in the template
}

// Process form submission to create a new gallery
func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID uint
		Title  string
	}
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")
	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, gallery, err)
		fmt.Println(err.Error()) // rudimentary logging
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	editGalleryPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editGalleryPath, http.StatusFound)
}
