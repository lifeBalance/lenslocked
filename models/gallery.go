package models

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Image struct {
	GalleryID int
	Path      string
	Filename  string
}

type Gallery struct {
	ID     int
	UserID uint
	Title  string
}

type GalleryService struct {
	DB *sql.DB
	// Folder to store images. If not set, defaults to "images".
	ImagesDir string
}

func (svc *GalleryService) Create(title string, userId uint) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userId,
	}
	row := svc.DB.QueryRow(`
		INSERT INTO galleries (title, user_id)
		VALUES ($1, $2)
		RETURNING id;
	`, gallery.Title, gallery.UserID)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery: %w", err)
	}
	return &gallery, nil
}

func (svc *GalleryService) GalleryById(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	row := svc.DB.QueryRow(`
		SELECT title, user_id
		FROM galleries
		WHERE id = $1;
	`, gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("gallery %w", ErrNotFound)
		}
		return nil, fmt.Errorf("query gallery by ID: %w", err)
	}
	return &gallery, nil
}

func (svc *GalleryService) GalleriesByUserId(userId uint) ([]Gallery, error) {
	rows, err := svc.DB.Query(`
		SELECT id, title
		FROM galleries
		WHERE user_id = $1;
	`, userId)
	if err != nil {
		return nil, fmt.Errorf("query galleries by user ID: %w", err)
	}
	var galleries []Gallery
	for rows.Next() {
		gallery := Gallery{
			UserID: userId,
		}
		err := rows.Scan(&gallery.ID, &gallery.Title)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user ID: %w", err)
		}
		galleries = append(galleries, gallery)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("query galleries by user ID: %w", err)
	}
	return galleries, nil
}

func (svc *GalleryService) UpdateGallery(gallery *Gallery) error {
	_, err := svc.DB.Exec(`
		UPDATE galleries
		SET title = $2
		WHERE id = $1;
	`, gallery.ID, gallery.Title)
	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (svc *GalleryService) DeleteGallery(galleryId int) error {
	_, err := svc.DB.Exec(`
		DELETE FROM galleries
		WHERE id = $1;
	`, galleryId)
	if err != nil {
		return fmt.Errorf("delete gallery: %w", err)
	}
	return nil
}

func (svc *GalleryService) Images(galleryId int) ([]Image, error) {
	globPattern := filepath.Join(svc.galleryDir(galleryId), "*") // "images/gallery-2/*"
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}
	var images []Image
	supportedExt := svc.supportedExtensions()
	for _, filename := range allFiles {
		if hasExtension(filename, supportedExt) {
			images = append(images, Image{
				GalleryID: galleryId,
				Path:      filename,
				Filename:  filepath.Base(filename),
			})
		}
	}
	return images, nil
}

func (svc *GalleryService) Image(galleryId int, filename string) (Image, error) {
	imgPath := filepath.Join(svc.galleryDir(galleryId), filename)
	_, err := os.Stat(imgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying single image: %w", err)
	}

	return Image{
		Filename:  filename,
		GalleryID: galleryId,
		Path:      imgPath,
	}, nil
}

func (svc *GalleryService) supportedExtensions() []string {
	// TODO: Set up list of supported extensions in .env or config.
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

func hasExtension(filename string, extensions []string) bool {
	for _, ext := range extensions {
		lowercasedFilename := strings.ToLower(filename)
		lowercasedExt := strings.ToLower(ext)
		if filepath.Ext(lowercasedFilename) == lowercasedExt {
			return true
		}
	}
	return false
}

func (svc *GalleryService) galleryDir(galleryId int) string {
	imagesDir := svc.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", galleryId))
}
