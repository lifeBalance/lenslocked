package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lifebalance/lenslocked/context"
	"github.com/lifebalance/lenslocked/models"
)

type Galleries struct {
	Templates struct {
		New   Template
		Index Template
		Show  Template
		Edit  Template
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

// Render form to edit a gallery
func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return
	}
	gallery, err := g.GalleryService.GalleryById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "gallery not found", http.StatusNotFound)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "you can't edit this gallery", http.StatusForbidden)
		return
	}
	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	g.Templates.Edit.Execute(w, r, data) // render title in the template
}

// Process form submission to edit a gallery
func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return
	}
	gallery, err := g.GalleryService.GalleryById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "gallery not found", http.StatusNotFound)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "you can't edit this gallery", http.StatusForbidden)
		return
	}

	gallery.Title = r.FormValue("title")
	err = g.GalleryService.UpdateGallery(gallery)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
	}
	var data struct {
		Galleries []Gallery
	}

	user := context.User(r.Context())
	galleries, err := g.GalleryService.GalleriesByUserId(user.ID)
	if err != nil {
		fmt.Println("galleries controller: index: ", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	for _, g := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID:    g.ID,
			Title: g.Title,
		})
	}
	g.Templates.Index.Execute(w, r, data)

}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return
	}
	gallery, err := g.GalleryService.GalleryById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "gallery not found", http.StatusNotFound)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	var mockGallery []string
	for range 20 {
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		catImgUrl := fmt.Sprintf("https://picsum.photos/%d/%d", w, h)
		mockGallery = append(mockGallery, catImgUrl)
	}
	data := struct {
		ID     int
		Title  string
		Images []string
	}{
		ID:     gallery.ID,
		Title:  gallery.Title,
		Images: mockGallery,
	}
	g.Templates.Show.Execute(w, r, data)
}
