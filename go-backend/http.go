package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type app struct {
	store    *Store
	cfg      config
	sessions *sessionStore
}

func newApp(store *Store, cfg config) *app {
	return &app{store: store, cfg: cfg, sessions: &sessionStore{sessions: make(map[string]adminSession)}}
}

// Keep the existing body size limit and reject trailing JSON values.
func decode(c *gin.Context, target any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(target); err != nil {
		c.String(http.StatusBadRequest, "Invalid JSON\n")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		c.String(http.StatusBadRequest, "Invalid JSON\n")
		return false
	}
	return true
}

func serverError(c *gin.Context, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		c.String(http.StatusNotFound, "Not found\n")
		return
	}
	log.Printf("request failed: %v", err)
	c.String(http.StatusInternalServerError, "Internal server error\n")
}

// Only public GET requests are allowed cross-origin; admin uses same-origin cookies.
func (a *app) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin := c.GetHeader("Origin"); origin != "" && strings.HasPrefix(c.Request.URL.Path, "/api/") {
			for _, allowed := range a.cfg.AllowedOrigins {
				if origin == strings.TrimSpace(allowed) {
					c.Header("Access-Control-Allow-Origin", origin)
					c.Writer.Header().Add("Vary", "Origin")
					if c.Request.Method == http.MethodOptions && c.GetHeader("Access-Control-Request-Method") == "GET" {
						c.Header("Access-Control-Allow-Methods", "GET")
						c.AbortWithStatus(http.StatusNoContent)
						return
					}
					break
				}
			}
		}
		c.Next()
	}
}

func (a *app) routes() http.Handler {
	router := gin.New()
	// Gin's default debug recovery logs request headers, which may contain session cookies.
	router.Use(gin.CustomRecoveryWithWriter(nil, func(c *gin.Context, _ any) {
		log.Printf("panic while handling %s", c.FullPath())
		c.AbortWithStatus(http.StatusInternalServerError)
	}), a.cors())
	router.HandleMethodNotAllowed = true
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false
	// Authentication inspects the immediate peer itself; do not trust arbitrary proxy headers.
	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	api := router.Group("/api")
	api.GET("/portfolio", a.portfolio)
	api.GET("/profile", a.profile)
	api.GET("/projects", a.projects)
	api.GET("/education", a.education)
	api.GET("/uploads/college/:filename", a.serveImage("college"))
	api.GET("/uploads/company-logos/:filename", a.serveImage("company-logos"))
	api.GET("/experience", a.experience)

	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Next()
	})
	admin.POST("/login", a.login)
	protected := admin.Group("")
	protected.Use(a.requireAdmin())
	protected.GET("/session", a.sessionInfo)
	protected.POST("/logout", a.logout)
	protected.PUT("/profile", a.updateProfile)
	protected.PUT("/education-photo", a.uploadEducationPhoto)
	protected.POST("/projects", a.saveProject)
	protected.PUT("/projects/:id", a.saveProject)
	protected.DELETE("/projects/:id", a.deleteItem("projects"))
	protected.POST("/education", a.saveEducation)
	protected.PUT("/education/:id", a.saveEducation)
	protected.DELETE("/education/:id", a.deleteItem("education"))
	protected.PUT("/experience/:id/logo", a.uploadCompanyLogo)
	protected.POST("/experience", a.saveExperience)
	protected.PUT("/experience/:id", a.saveExperience)
	protected.DELETE("/experience/:id", a.deleteItem("experience"))
	return router
}

func (a *app) portfolio(c *gin.Context) {
	data, err := a.store.portfolio(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (a *app) profile(c *gin.Context) {
	data, err := a.store.profile(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (a *app) projects(c *gin.Context) {
	data, err := a.store.projects(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (a *app) education(c *gin.Context) {
	data, err := a.store.education(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (a *app) experience(c *gin.Context) {
	data, err := a.store.experiences(c.Request.Context())
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func required(value **string) bool {
	if *value == nil || strings.TrimSpace(**value) == "" {
		return false
	}
	trimmed := strings.TrimSpace(**value)
	*value = &trimmed
	return true
}

func (a *app) updateProfile(c *gin.Context) {
	var p Profile
	if !decode(c, &p) {
		return
	}
	if !required(&p.Name) {
		c.String(http.StatusBadRequest, "Name is required\n")
		return
	}
	saved, err := a.store.saveProfile(c.Request.Context(), p)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, saved)
}

// A missing ID indicates a POST; PUT and DELETE routes always provide one.
func itemID(c *gin.Context) (int64, bool) {
	if c.Request.Method == http.MethodPost {
		return 0, true
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.String(http.StatusNotFound, "Not found\n")
		return 0, false
	}
	return id, true
}

func saveStatus(id int64) int {
	if id == 0 {
		return http.StatusCreated
	}
	return http.StatusOK
}

func (a *app) saveProject(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	var p Project
	if !decode(c, &p) {
		return
	}
	if !required(&p.Title) {
		c.String(http.StatusBadRequest, "Title is required\n")
		return
	}
	saved, err := a.store.saveProject(c.Request.Context(), p, id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(saveStatus(id), saved)
}

func (a *app) saveEducation(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	var e Education
	if !decode(c, &e) {
		return
	}
	if !required(&e.Institution) {
		c.String(http.StatusBadRequest, "Institution is required\n")
		return
	}
	saved, err := a.store.saveEducation(c.Request.Context(), e, id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(saveStatus(id), saved)
}

func (a *app) saveExperience(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	var e Experience
	if !decode(c, &e) {
		return
	}
	if !required(&e.Company) {
		c.String(http.StatusBadRequest, "Company is required\n")
		return
	}
	saved, err := a.store.saveExperience(c.Request.Context(), e, id)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(saveStatus(id), saved)
}

func (a *app) deleteItem(kind string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := itemID(c)
		if !ok {
			return
		}
		if err := a.store.deleteItem(c.Request.Context(), kind, id); err != nil {
			serverError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}
