package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetOffset_WithValidOffset(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		offset, err := getOffset(c)
		assert.NoError(t, err)
		assert.Equal(t, 10, offset)
		c.JSON(http.StatusOK, gin.H{"offset": offset})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?offset=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOffset_WithoutOffset(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		offset, err := getOffset(c)
		assert.NoError(t, err)
		assert.Equal(t, 0, offset)
		c.JSON(http.StatusOK, gin.H{"offset": offset})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetOffset_WithInvalidOffset(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		offset, err := getOffset(c)
		assert.Error(t, err)
		assert.Equal(t, 0, offset)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?offset=invalid", nil)
	router.ServeHTTP(w, req)
}

func TestGetOffset_WithZeroOffset(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		offset, err := getOffset(c)
		assert.NoError(t, err)
		assert.Equal(t, 0, offset)
		c.JSON(http.StatusOK, gin.H{"offset": offset})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?offset=0", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRespondSuccess(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		data := gin.H{"key": "value"}
		respondSuccess(c, http.StatusOK, data)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "key")
	assert.Contains(t, w.Body.String(), "value")
}

func TestRespondError(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		testErr := assert.AnError
		respondError(c, http.StatusBadRequest, testErr)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Err")
}
