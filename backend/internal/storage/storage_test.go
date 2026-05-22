package storage

import (
	"archive/zip"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestUnzipImagesRejectsPathTraversal(t *testing.T) {
	storage := New(t.TempDir())
	projectID := uuid.New()
	if err := storage.EnsureProjectDirs(projectID); err != nil {
		t.Fatalf("EnsureProjectDirs: %v", err)
	}
	archivePath := filepath.Join(t.TempDir(), "unsafe.zip")
	writeZip(t, archivePath, map[string][]byte{"../evil.png": []byte("nope")})

	if err := storage.UnzipImages(projectID, archivePath); err == nil {
		t.Fatal("UnzipImages accepted path traversal entry")
	}
}

func TestSafeImagePathSupportsNestedImagesAndFallbackPrefix(t *testing.T) {
	storage := New(t.TempDir())
	projectID := uuid.New()
	if err := storage.EnsureProjectDirs(projectID); err != nil {
		t.Fatalf("EnsureProjectDirs: %v", err)
	}
	archivePath := filepath.Join(t.TempDir(), "images.zip")
	writeZip(t, archivePath, map[string][]byte{"cat.png": tinyPNG(t)})
	if err := storage.UnzipImages(projectID, archivePath); err != nil {
		t.Fatalf("UnzipImages: %v", err)
	}

	if !storage.ImageExists(projectID, "images/cat.png") {
		t.Fatal("ImageExists should accept CSV paths prefixed with images/")
	}
	if !storage.ImageReadable(projectID, "images/cat.png") {
		t.Fatal("ImageReadable should decode the PNG")
	}
	if _, err := storage.SafeImagePath(projectID, "../cat.png"); err == nil {
		t.Fatal("SafeImagePath accepted traversal")
	}
}

func TestSafeImagePathSupportsZipImagesFolder(t *testing.T) {
	storage := New(t.TempDir())
	projectID := uuid.New()
	if err := storage.EnsureProjectDirs(projectID); err != nil {
		t.Fatalf("EnsureProjectDirs: %v", err)
	}
	archivePath := filepath.Join(t.TempDir(), "images.zip")
	writeZip(t, archivePath, map[string][]byte{"images/cat.png": tinyPNG(t)})
	if err := storage.UnzipImages(projectID, archivePath); err != nil {
		t.Fatalf("UnzipImages: %v", err)
	}

	if !storage.ImageExists(projectID, "cat.png") {
		t.Fatal("ImageExists should accept archives with images/ top-level folder")
	}
}

func TestSafeImagePathSupportsAnimals10RawArchive(t *testing.T) {
	storage := New(t.TempDir())
	projectID := uuid.New()
	if err := storage.EnsureProjectDirs(projectID); err != nil {
		t.Fatalf("EnsureProjectDirs: %v", err)
	}
	archivePath := filepath.Join(t.TempDir(), "images.zip")
	writeZip(t, archivePath, map[string][]byte{"Animals-10/butterfly/butterfly (1).jpeg": tinyPNG(t)})
	if err := storage.UnzipImages(projectID, archivePath); err != nil {
		t.Fatalf("UnzipImages: %v", err)
	}

	if !storage.ImageExists(projectID, "butterfly_000001.jpeg") {
		t.Fatal("ImageExists should map prepared Animals10 names to raw Animals-10 archive names")
	}
	if err := storage.EnsureImageAliases(projectID, []string{"butterfly_000001.jpeg"}); err != nil {
		t.Fatalf("EnsureImageAliases: %v", err)
	}
	if !storage.ImageExists(projectID, "butterfly_000001.jpeg") {
		t.Fatal("ImageExists should still work after alias creation")
	}
	if _, err := os.Stat(filepath.Join(storage.ImagesDir(projectID), "butterfly_000001.jpeg")); err != nil {
		t.Fatalf("alias was not created: %v", err)
	}
}

func writeZip(t *testing.T, path string, files map[string][]byte) {
	t.Helper()
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	writer := zip.NewWriter(out)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
}

func tinyPNG(t *testing.T) []byte {
	t.Helper()
	bytes, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADUlEQVR4nGJgYGD4DwABBAEAgh7e1wAAAABJRU5ErkJggg==")
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	return bytes
}
