package storage

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Storage struct {
	root string
}

func New(root string) *Storage {
	return &Storage{root: root}
}

func (s *Storage) Health() error {
	return os.MkdirAll(s.root, 0o755)
}

func (s *Storage) ProjectDir(projectID uuid.UUID) string {
	return filepath.Join(s.root, "projects", projectID.String())
}

func (s *Storage) RawDir(projectID uuid.UUID) string {
	return filepath.Join(s.ProjectDir(projectID), "raw")
}

func (s *Storage) ImagesDir(projectID uuid.UUID) string {
	return filepath.Join(s.RawDir(projectID), "images")
}

func (s *Storage) AnalysisDir(projectID uuid.UUID) string {
	return filepath.Join(s.ProjectDir(projectID), "analysis")
}

func (s *Storage) ExportsDir(projectID uuid.UUID) string {
	return filepath.Join(s.ProjectDir(projectID), "exports")
}

func (s *Storage) EnsureProjectDirs(projectID uuid.UUID) error {
	for _, dir := range []string{s.RawDir(projectID), s.ImagesDir(projectID), s.AnalysisDir(projectID), s.ExportsDir(projectID)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) SaveMultipartFile(projectID uuid.UUID, file multipart.File, filename string) (string, error) {
	if err := s.EnsureProjectDirs(projectID); err != nil {
		return "", err
	}
	name := filepath.Base(filename)
	if name == "." || name == string(filepath.Separator) {
		return "", errors.New("invalid filename")
	}
	path := filepath.Join(s.RawDir(projectID), name)
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Storage) UnzipImages(projectID uuid.UUID, archivePath string) error {
	if err := os.RemoveAll(s.ImagesDir(projectID)); err != nil {
		return err
	}
	if err := os.MkdirAll(s.ImagesDir(projectID), 0o755); err != nil {
		return err
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	targetRoot, err := filepath.Abs(s.ImagesDir(projectID))
	if err != nil {
		return err
	}
	for _, file := range reader.File {
		cleanName := filepath.Clean(file.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return fmt.Errorf("unsafe zip path: %s", file.Name)
		}
		targetPath := filepath.Join(targetRoot, cleanName)
		if !strings.HasPrefix(targetPath, targetRoot+string(os.PathSeparator)) && targetPath != targetRoot {
			return fmt.Errorf("unsafe zip path: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if !file.FileInfo().Mode().IsRegular() {
			return fmt.Errorf("unsupported zip entry: %s", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.Create(targetPath)
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := errors.Join(src.Close(), dst.Close())
		if copyErr != nil || closeErr != nil {
			return errors.Join(copyErr, closeErr)
		}
	}
	return nil
}

func (s *Storage) ImageExists(projectID uuid.UUID, relativePath string) bool {
	_, err := s.SafeImagePath(projectID, relativePath)
	return err == nil
}

func (s *Storage) ImageReadable(projectID uuid.UUID, relativePath string) bool {
	path, err := s.SafeImagePath(projectID, relativePath)
	if err != nil {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".webp" {
		info, err := os.Stat(path)
		return err == nil && info.Size() > 0
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
	return err == nil && config.Width > 0 && config.Height > 0
}

func (s *Storage) FileSHA256(projectID uuid.UUID, relativePath string) (string, error) {
	path, err := s.SafeImagePath(projectID, relativePath)
	if err != nil {
		return "", err
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *Storage) SafeImagePath(projectID uuid.UUID, relativePath string) (string, error) {
	if relativePath == "" {
		return "", errors.New("empty image path")
	}
	if !s.IsSafeRelativePath(relativePath) {
		return "", fmt.Errorf("unsafe image path: %s", relativePath)
	}

	clean := filepath.Clean(relativePath)
	candidates := []string{filepath.Join(s.ImagesDir(projectID), clean)}
	prefix := "images" + string(filepath.Separator)
	if strings.HasPrefix(clean, prefix) {
		candidates = append(candidates, filepath.Join(s.ImagesDir(projectID), strings.TrimPrefix(clean, prefix)))
	}
	root, err := filepath.Abs(s.ImagesDir(projectID))
	if err != nil {
		return "", err
	}
	for _, candidate := range candidates {
		absCandidate, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absCandidate, root+string(os.PathSeparator)) && absCandidate != root {
			continue
		}
		info, err := os.Stat(absCandidate)
		if err == nil && !info.IsDir() {
			return absCandidate, nil
		}
	}
	return "", os.ErrNotExist
}

func (s *Storage) IsSafeRelativePath(relativePath string) bool {
	if relativePath == "" || filepath.IsAbs(relativePath) {
		return false
	}
	clean := filepath.Clean(relativePath)
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func (s *Storage) HasSupportedImageExtension(relativePath string) bool {
	switch strings.ToLower(filepath.Ext(relativePath)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func (s *Storage) ImageContentType(path string) string {
	if value := mime.TypeByExtension(strings.ToLower(filepath.Ext(path))); value != "" {
		return value
	}
	return "application/octet-stream"
}
