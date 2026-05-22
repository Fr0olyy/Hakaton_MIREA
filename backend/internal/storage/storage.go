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
	"strconv"
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

func (s *Storage) DatasetPath(projectID uuid.UUID) string {
	return filepath.Join(s.RawDir(projectID), "dataset.csv")
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
	candidates := s.imagePathCandidates(projectID, clean)
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

func (s *Storage) EnsureImageAliases(projectID uuid.UUID, relativePaths []string) error {
	for _, relativePath := range relativePaths {
		if relativePath == "" || !s.IsSafeRelativePath(relativePath) {
			continue
		}
		clean := filepath.Clean(relativePath)
		directPath := filepath.Join(s.ImagesDir(projectID), clean)
		if _, err := os.Stat(directPath); err == nil {
			continue
		}
		sourcePath, err := s.SafeImagePath(projectID, clean)
		if err != nil || sourcePath == directPath {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(directPath), 0o755); err != nil {
			return err
		}
		if err := os.Link(sourcePath, directPath); err == nil {
			continue
		}
		if err := copyFile(directPath, sourcePath); err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) imagePathCandidates(projectID uuid.UUID, clean string) []string {
	root := s.ImagesDir(projectID)
	candidates := []string{}
	add := func(parts ...string) {
		candidate := filepath.Join(append([]string{root}, parts...)...)
		for _, existing := range candidates {
			if existing == candidate {
				return
			}
		}
		candidates = append(candidates, candidate)
	}

	add(clean)
	prefix := "images" + string(filepath.Separator)
	if strings.HasPrefix(clean, prefix) {
		add(strings.TrimPrefix(clean, prefix))
	} else {
		add("images", clean)
	}
	for _, candidate := range animals10Candidates(clean) {
		add(candidate)
	}
	return candidates
}

func animals10Candidates(clean string) []string {
	base := filepath.Base(clean)
	ext := strings.ToLower(filepath.Ext(base))
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	separator := strings.LastIndex(stem, "_")
	if separator <= 0 || separator == len(stem)-1 {
		return nil
	}
	className := stem[:separator]
	numberText := strings.TrimLeft(stem[separator+1:], "0")
	if numberText == "" {
		numberText = "0"
	}
	number, err := strconv.Atoi(numberText)
	if err != nil {
		return nil
	}
	extensions := uniqueStrings([]string{ext, ".jpg", ".jpeg", ".png", ".webp"})
	candidates := make([]string, 0, len(extensions)*3)
	for _, extension := range extensions {
		if extension == "" {
			continue
		}
		fileName := fmt.Sprintf("%s (%d)%s", className, number, extension)
		candidates = append(candidates,
			filepath.Join(className, fileName),
			filepath.Join("Animals-10", className, fileName),
			filepath.Join("animals10", className, fileName),
		)
	}
	return candidates
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

func copyFile(target, source string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	return errors.Join(copyErr, closeErr)
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		seen := false
		for _, existing := range result {
			if existing == value {
				seen = true
				break
			}
		}
		if !seen {
			result = append(result, value)
		}
	}
	return result
}
