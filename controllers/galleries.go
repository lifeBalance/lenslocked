package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

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

type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

// New renders the form to create a new gallery.
//
// It uses the `title` query param (if present) to pre-fill the form.
func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")   // parse query string
	g.Templates.New.Execute(w, r, data) // render title in the template
}

// Create processes the form submission to create a new gallery.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery creation fails (500)
//
// On success, it redirects to `/galleries/{id}/edit`.
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

// Index lists the current user's galleries.
//
// It writes an HTTP error response and returns early when:
//
// - loading galleries fails (500)
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

// Show displays a gallery and its images.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the gallery doesn't exist (404)
//
// - loading images fails (500)
func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r)
	if err != nil {
		return
	}

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}
	// data for the template
	data := struct {
		ID     int
		Title  string
		Images []Image
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	for _, img := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       gallery.ID,
			Filename:        img.Filename,
			FilenameEscaped: url.PathEscape(img.Filename),
		})
	}

	g.Templates.Show.Execute(w, r, data)
}

// Edit renders the form to edit a gallery.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the gallery doesn't exist (404)
//
// - the current user does not own the gallery (403)
//
// - loading images fails (500)
func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	data := struct {
		ID     int
		Title  string
		Images []Image
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	// Attach the images to the data
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	for _, img := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       gallery.ID,
			Filename:        img.Filename,
			FilenameEscaped: url.PathEscape(img.Filename),
		})
	}
	g.Templates.Edit.Execute(w, r, data) // render title in the template
}

// Update processes the form submission to update a gallery.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the gallery doesn't exist (404)
//
// - the current user does not own the gallery (403)
//
// - updating the gallery fails (500)
//
// On success, it redirects back to `/galleries/{id}/edit`.
func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r, userMustOwnGallery)
	if err != nil {
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

func (g Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = r.ParseMultipartForm(5 << 20) // 5mb (shift x10 = 1kb)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	fileHeaders := r.MultipartForm.File["images"]
	for _, fh := range fileHeaders {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		err = g.GalleryService.CreateImage(gallery.ID, fh.Filename, file)
		if err != nil {
			var fileErr models.FileError
			if errors.As(err, &fileErr) {
				msg := fmt.Sprintf("%v has invalid content type or extension (only allowed; jpeg/jpg, png, and gif)", fh.Filename)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) UploadImageViaURL(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	files := r.PostForm["files"]
	var wg sync.WaitGroup
	wg.Add(len(files))
	for _, f := range files {
		go func(imageUrl string) {
			err = g.GalleryService.CreateImageViaURL(gallery.ID, imageUrl)
			if err != nil {
				http.Error(w, "something went wrong creating image: "+f, http.StatusInternalServerError)
			}
			wg.Done()
		}(f) // 👈 Pass f here! ⚠️
	}
	wg.Wait()
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

// Delete deletes a gallery by its `id` URL param.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the gallery doesn't exist (404)
//
// - deleting the gallery fails (500)
//
// On success, it redirects to `/galleries`.
func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryById(w, r)
	if err != nil {
		return
	}
	err = g.GalleryService.DeleteGallery(gallery.ID)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

// DeleteImage deletes an image from a gallery.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the gallery doesn't exist (404)
//
// - the current user does not own the gallery (403)
//
// - deleting the image fails (500)
//
// On success, it redirects back to `/galleries/{id}/edit`.
func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	// filename := chi.URLParam(r, "filename")
	filename := g.filename(w, r) // fix traversal attacks

	gallery, err := g.galleryById(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

// Image serves a single image file from a gallery.
//
// It writes an HTTP error response and returns early when:
//
// - the gallery `id` URL param is invalid (404)
//
// - the image doesn't exist (404)
//
// - loading the image fails for other reasons (500)
func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	// filename := chi.URLParam(r, "filename")
	filename := g.filename(w, r) // fix traversal attacks

	galleryId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return
	}
	image, err := g.GalleryService.Image(galleryId, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "image not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, image.Path)
}

// galleryById loads and returns the gallery referenced by the route param `id`.
//
// In case of error, it returns it, and writes an HTTP error response when:
//
// - the `id` URL param is not an int (404)
//
// - the gallery doesn't exist (404)
//
// - the lookup fails for other reasons (500)
//
// It also runs any `galleryOpt` callbacks after loading the gallery.
func (g Galleries) galleryById(
	w http.ResponseWriter,
	r *http.Request,
	opts ...galleryOpt,
) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.GalleryService.GalleryById(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "gallery not found", http.StatusNotFound)
			return nil, err

		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}
	// Run functional options.
	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}
	return gallery, nil
}

// userMustOwnGallery is a `galleryOpt` that requires the current user to own the gallery.
//
// It writes an HTTP error response and returns a non-nil error when:
//
// - the current user does not own the gallery (403)
func userMustOwnGallery(
	w http.ResponseWriter,
	r *http.Request,
	gallery *models.Gallery,
) error {
	user := context.User(r.Context())
	if gallery.UserID != user.ID {
		http.Error(w, "you can't edit this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}

func (g Galleries) filename(w http.ResponseWriter, r *http.Request) string {
	_ = w
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}
