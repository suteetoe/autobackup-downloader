package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	maxUploadSize = 100 << 20 // 100 MB
)

type AppConfig struct {
	BasePath       string
	FileStorageDir string
	Port           string
}

type FileInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
}

func main() {
	config := loadConfig()
	if err := os.MkdirAll(config.FileStorageDir, 0755); err != nil {
		panic(err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("100M"))

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, config.BasePath+"/")
	})

	app := e.Group(config.BasePath)
	app.Static("/", "public")

	api := app.Group("/api")
	api.GET("/files", listFiles(config.FileStorageDir, config.BasePath))
	api.POST("/files", uploadFile(config.FileStorageDir, config.BasePath))
	api.GET("/files/:name/download", downloadFile(config.FileStorageDir))
	api.DELETE("/files/:name", deleteFile(config.FileStorageDir))

	e.Logger.Printf("base path: %s", config.BasePath)
	e.Logger.Printf("file storage directory: %s", config.FileStorageDir)
	e.Logger.Fatal(e.Start(":" + config.Port))
}

func loadConfig() AppConfig {
	basePath := normalizeBasePath(os.Getenv("BASE_PATH"))
	if basePath == "" {
		basePath = "/autobackupfile"
	}

	fileStorageDir := os.Getenv("FILE_STORAGE_DIR")
	if fileStorageDir == "" {
		fileStorageDir = "files"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return AppConfig{
		BasePath:       basePath,
		FileStorageDir: fileStorageDir,
		Port:           port,
	}
}

func normalizeBasePath(basePath string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" || basePath == "/" {
		return ""
	}
	return "/" + strings.Trim(basePath, "/")
}

func listFiles(fileStorageDir string, basePath string) echo.HandlerFunc {
	return func(c echo.Context) error {
		entries, err := os.ReadDir(fileStorageDir)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot read file list")
		}

		files := make([]FileInfo, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := entry.Name()
			files = append(files, FileInfo{
				Name:      name,
				Size:      info.Size(),
				UpdatedAt: info.ModTime().Format(time.RFC3339),
				URL:       fmt.Sprintf("%s/api/files/%s/download", basePath, name),
			})
		}

		sort.Slice(files, func(i, j int) bool {
			return files[i].UpdatedAt > files[j].UpdatedAt
		})

		return c.JSON(http.StatusOK, files)
	}
}

func uploadFile(fileStorageDir string, basePath string) echo.HandlerFunc {
	return func(c echo.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "form field 'file' is required")
		}
		if file.Size > maxUploadSize {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "file is larger than 100 MB")
		}

		src, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot open uploaded file")
		}
		defer src.Close()

		filename, err := safeFilename(file.Filename)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		dstPath := filepath.Join(fileStorageDir, filename)
		dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if errors.Is(err, os.ErrExist) {
			return echo.NewHTTPError(http.StatusConflict, "file already exists")
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot create destination file")
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot save uploaded file")
		}

		return c.JSON(http.StatusCreated, map[string]string{
			"name": filename,
			"url":  fmt.Sprintf("%s/api/files/%s/download", basePath, filename),
		})
	}
}

func downloadFile(fileStorageDir string) echo.HandlerFunc {
	return func(c echo.Context) error {
		filename, err := safeFilename(c.Param("name"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		path := filepath.Join(fileStorageDir, filename)
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "file not found")
		}

		return c.Attachment(path, filename)
	}
}

func deleteFile(fileStorageDir string) echo.HandlerFunc {
	return func(c echo.Context) error {
		filename, err := safeFilename(c.Param("name"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		path := filepath.Join(fileStorageDir, filename)
		if err := os.Remove(path); errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusNotFound, "file not found")
		} else if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "cannot delete file")
		}

		return c.NoContent(http.StatusNoContent)
	}
}

func safeFilename(name string) (string, error) {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", errors.New("invalid filename")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", errors.New("filename cannot contain path separators")
	}
	return name, nil
}
