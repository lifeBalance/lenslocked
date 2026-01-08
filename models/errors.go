package models

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

var (
	ErrEmailTaken = errors.New("email address already taken")
	ErrNotFound   = errors.New("not found")
)

type FileError struct {
	Issue string
}

func (fe FileError) Error() string {
	return fmt.Sprintf("invalid file: %v", fe.Issue)
}

func checkContentType(r io.Reader, allowedTypes []string) ([]byte, error) {
	testBytes := make([]byte, 512)
	bytesRead, err := r.Read(testBytes)
	if err != nil {
		return nil, fmt.Errorf("checking content type: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("checking content type: %w", err)
	}
	contentType := http.DetectContentType(testBytes)
	for _, t := range allowedTypes {
		if contentType == t {
			return testBytes[:bytesRead], nil
		}
	}
	return nil, FileError{
		Issue: fmt.Sprintf("invalid content type: %v", contentType),
	}
}

func checkExtension(filename string, allowedExtensions []string) error {
	if hasExtension(filename, allowedExtensions) {
		return nil
	}
	return FileError{
		Issue: fmt.Sprintf("invalid extension: %v", filepath.Ext(filename)),
	}
}
