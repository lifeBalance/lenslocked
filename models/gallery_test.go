package models

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// This test demonstrates a path traversal issue in GalleryService.Image.
//
// It hardcodes the gallery IDs:  attack from gallery 3 → target gallery 2
// and depends on the fixture image `images/gallery-2/146-397x363.jpg`;
// keep that file available when running the test.
func TestGalleryServiceImagePathTraversal(t *testing.T) {
	t.Helper()

	const (
		repoImagesDir     = "../images"
		attackerGalleryID = 3
		targetGalleryDir  = "gallery-2"
		targetImage       = "146-397x363.jpg"
	)

	targetFullPath := filepath.Join(repoImagesDir, targetGalleryDir, targetImage)
	if _, err := os.Stat(targetFullPath); err != nil {
		t.Fatalf("fixture image missing: %v", err)
	}

	svc := &GalleryService{ImagesDir: repoImagesDir}

	// Attempt to escape "gallery-3" into "gallery-2".
	traversalExploit := fmt.Sprintf("../%s/%s", targetGalleryDir, targetImage)
	_, err := svc.Image(attackerGalleryID, traversalExploit)
	if err == nil {
		t.Fatalf("expected traversal to be rejected")
	}
}
