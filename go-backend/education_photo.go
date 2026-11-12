package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
)

var imageFilename = regexp.MustCompile(`^[a-f0-9]{64}\.(jpeg|png)$`)

func (a *app) serveImage(dir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("filename")
		if !imageFilename.MatchString(name) {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.File(filepath.Join(a.cfg.UploadDir, dir, name))
	}
}

func (a *app) saveUploadedImage(c *gin.Context, folder string) (string, bool) {
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20))
	if err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			c.String(http.StatusRequestEntityTooLarge, "Choose a photo of 5 MiB or less.")
		} else {
			c.String(http.StatusBadRequest, "Could not read the upload.")
		}
		return "", false
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") {
		c.String(http.StatusUnsupportedMediaType, "Choose a JPEG or PNG photo.")
		return "", false
	}
	// Bound the decoded allocation as well as the compressed upload.
	if cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 24000000/cfg.Height {
		c.String(http.StatusBadRequest, "Choose a photo of 24 megapixels or less.")
		return "", false
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		c.String(http.StatusBadRequest, "This photo could not be decoded.")
		return "", false
	}
	dir := filepath.Join(a.cfg.UploadDir, folder)
	name := fmt.Sprintf("%x.%s", sha256.Sum256(data), format)
	if err := writeUpload(dir, name, data); err != nil {
		serverError(c, err)
		return "", false
	}
	return "/api/uploads/" + folder + "/" + name, true
}

func (a *app) uploadEducationPhoto(c *gin.Context) {
	path, ok := a.saveUploadedImage(c, "college")
	if !ok {
		return
	}
	_, err := a.store.db.ExecContext(c.Request.Context(), `INSERT INTO education_photo (id, image_path)
		VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET image_path=EXCLUDED.image_path`, path)
	if err != nil {
		serverError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *app) uploadCompanyLogo(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	var entryID int64
	err := a.store.db.QueryRowContext(ctx, "SELECT id FROM experience WHERE id=$1", id).Scan(&entryID)
	if err != nil {
		serverError(c, err)
		return
	}
	path, ok := a.saveUploadedImage(c, "company-logos")
	if !ok {
		return
	}
	err = a.store.db.QueryRowContext(ctx, "UPDATE experience SET company_logo_path=$1 WHERE id=$2 RETURNING id", path, id).Scan(&entryID)
	if err != nil {
		serverError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Publish the complete file before making its URL current in PostgreSQL.
func writeUpload(dir, name string, data []byte) error {
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filepath.Join(dir, name))
}
