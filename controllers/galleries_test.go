package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/lifebalance/lenslocked/models"
)

// go test ./controllers -run TestGalleriesImageRejectsTraversal
func TestGalleriesImageRejectsTraversal(t *testing.T) {
	t.Helper()

	tempImages := t.TempDir()
	gallery2Dir := filepath.Join(tempImages, "gallery-2")
	gallery3Dir := filepath.Join(tempImages, "gallery-3")

	if err := os.MkdirAll(gallery2Dir, 0o755); err != nil {
		t.Fatalf("mkdir gallery-2: %v", err)
	}
	if err := os.MkdirAll(gallery3Dir, 0o755); err != nil {
		t.Fatalf("mkdir gallery-3: %v", err)
	}

	secretImage := "146-397x363.jpg"
	if err := os.WriteFile(filepath.Join(gallery2Dir, secretImage), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write secret image: %v", err)
	}

	galleryService := &models.GalleryService{ImagesDir: tempImages}
	ctrl := Galleries{GalleryService: galleryService}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/galleries/3/images/../gallery-2/"+secretImage, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "3")
	rctx.URLParams.Add("filename", "../gallery-2/"+secretImage)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	ctrl.Image(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for traversal attempt, got %d", rr.Code)
	}
}
