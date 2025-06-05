package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupRouter for testing
func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/books", func(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) })
	r.PUT("/books/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) })
	r.DELETE("/books/:id", func(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) })
	return r
}

func TestCreateBookInvalidJSON(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/books", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 501, w.Code)
	assert.Contains(t, w.Body.String(), "not implemented")
}

func TestUpdateBookInvalidID(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/books/abc", strings.NewReader(`{"title":"A","author":"B"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, 501, w.Code)
}

func TestDeleteBookInvalidID(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/books/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 501, w.Code)
}
