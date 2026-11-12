package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func (a *app) uploadResume(c *gin.Context) {
	data, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, 5<<20))
	if err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			c.String(http.StatusRequestEntityTooLarge, "Choose a PDF of 5 MiB or less.")
		} else {
			c.String(http.StatusBadRequest, "Could not read the upload.")
		}
		return
	}
	if http.DetectContentType(data) != "application/pdf" || !bytes.HasSuffix(bytes.TrimSpace(data), []byte("%%EOF")) {
		c.String(http.StatusUnsupportedMediaType, "Choose a complete PDF file.")
		return
	}

	name := fmt.Sprintf("%x.pdf", sha256.Sum256(data))
	if err := writeUpload(filepath.Join(a.cfg.UploadDir, "resumes"), name, data); err != nil {
		serverError(c, err)
		return
	}
	var id int64
	err = a.store.db.QueryRowContext(c.Request.Context(), `UPDATE profile SET resume_path=$1
		WHERE id=(SELECT id FROM profile ORDER BY id LIMIT 1) RETURNING id`, name).Scan(&id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (a *app) serveResume(c *gin.Context) {
	var name *string
	err := a.store.db.QueryRowContext(c.Request.Context(), `SELECT resume_path FROM profile ORDER BY id LIMIT 1`).Scan(&name)
	if err != nil {
		serverError(c, err)
		return
	}
	if name == nil {
		c.Status(http.StatusNotFound)
		return
	}
	disposition := "inline"
	if c.Query("download") == "1" {
		disposition = "attachment"
	}
	c.Header("Content-Disposition", disposition+`; filename="resume.pdf"`)
	c.Header("Content-Type", "application/pdf")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "no-store")
	c.File(filepath.Join(a.cfg.UploadDir, "resumes", filepath.Base(*name)))
}
