package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
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

// Create creates a new gallery for a user.
//
// It writes to the database and returns the created gallery (including its ID).
//
// It returns a non-nil error when:
//
// - inserting the gallery fails
//
// - scanning the inserted ID fails
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

func (svc *GalleryService) CreateImage(
	galleryId int,
	filename string,
	fileContents io.Reader,
) error {
	galleryDir := svc.galleryDir(galleryId)
	err := os.MkdirAll(galleryDir, 0755)
	if err != nil {
		return fmt.Errorf("create image folder: %w", err)
	}
	imagePath := filepath.Join(galleryDir, filename)
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer file.Close()
	_, err = io.Copy(file, fileContents)
	if err != nil {
		return fmt.Errorf("copying image to file: %w", err)
	}
	return nil
}

// GalleryById queries a single gallery by its ID.
//
// It returns a non-nil error when:
//
// - the gallery does not exist (wraps `ErrNotFound`)
//
// - the database query fails for other reasons
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

// GalleriesByUserId queries all galleries owned by a user.
//
// It returns a non-nil error when:
//
// - the database query fails
//
// - scanning any row fails
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

// UpdateGallery updates a gallery's mutable fields.
//
// It currently updates only the `title`.
//
// It returns a non-nil error when:
//
// - the database update fails
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

// DeleteGallery deletes a gallery by ID.
//
// It returns a non-nil error when:
//
// - the database delete fails
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

// Images lists image files stored for a gallery.
//
// It filters the files on disk using `supportedExtensions`.
//
// It returns a non-nil error when:
//
// - retrieving the list of files (glob) fails
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

// Image returns a single image from a gallery.
//
// It returns a non-nil error when:
//
// - the image file does not exist (returns `ErrNotFound`)
//
// - checking the file fails for other reasons
func (svc *GalleryService) Image(galleryId int, filename string) (Image, error) {
	galleryDir := svc.galleryDir(galleryId)

	sanitizedPath, err := sanitizedGalleryPath(galleryDir, filename)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying single image: %w", err)
	}

	_, err = os.Stat(sanitizedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying single image: %w", err)
	}

	return Image{
		Filename:  filepath.Base(sanitizedPath),
		GalleryID: galleryId,
		Path:      sanitizedPath,
	}, nil
}

// DeleteImage deletes an image from a gallery.
//
// It returns a non-nil error when:
//
// - the image does not exist (wraps `ErrNotFound`)
//
// - removing the file fails
func (svc *GalleryService) DeleteImage(galleryId int, filename string) error {
	img, err := svc.Image(galleryId, filename)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	err = os.Remove(img.Path)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}
	return nil
}

// sanitizedGalleryPath ensures filename resolves within galleryDir.
func sanitizedGalleryPath(galleryDir, filename string) (string, error) {
	joinedPath := filepath.Join(galleryDir, filename)
	cleanPath := filepath.Clean(joinedPath)

	absGalleryDir, err := filepath.Abs(galleryDir)
	if err != nil {
		return "", err
	}
	absImagePath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absGalleryDir, absImagePath)
	if err != nil {
		return "", err
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", ErrNotFound
	}

	return absImagePath, nil
}

// supportedExtensions lists file extensions supported as gallery images.
func (svc *GalleryService) supportedExtensions() []string {
	// TODO: Set up list of supported extensions in .env or config.
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}

// hasExtension reports whether filename has one of the provided extensions.
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

// galleryDir returns the folder path used to store a gallery's images.
//
// If `ImagesDir` is not set, it defaults to `images`.
func (svc *GalleryService) galleryDir(galleryId int) string {
	imagesDir := svc.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", galleryId))
}
