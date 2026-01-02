package models

import (
	"database/sql"
	"errors"
	"fmt"
)

type Gallery struct {
	ID     int
	UserID uint
	Title  string
}

type GalleryService struct {
	DB *sql.DB
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
