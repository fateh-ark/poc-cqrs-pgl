package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mock implementations and helpers would go here

func setupRouter() *gin.Engine {
	// Use a test router with the same routes
	r := gin.Default()
	r.GET("/books/:id", getBook)
	r.GET("/books", listBooks)
	return r
}

func TestGetBookInvalidID(t *testing.T) {
	r := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/books/abc", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid ID")
}

// More tests would be added here for DB/RabbitMQ mocking
