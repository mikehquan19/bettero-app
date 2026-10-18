package controllers

import (
	"betterov2/models"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// respondSuccess sends response to API request that successfully executed
func respondSuccess(c *gin.Context, successCode int, data any) {
	c.JSON(successCode, models.Response{Data: data})
}

// respondError sends response to API request that is not succesfully executed
func respondError(c *gin.Context, errorCode int, err error) {
	log.Println(err)
	c.JSON(errorCode, models.Response{Err: err.Error()})
}

func respondAuthInternal(c *gin.Context, err error) {
	log.Printf("authentication request failed: %v", err)
	c.JSON(http.StatusInternalServerError, models.Response{Err: "internal server error"})
}

func isAuthenticationError(err error) bool {
	return errors.Is(err, models.ErrInvalidCredentials) || errors.Is(err, models.ErrNotFound)
}

func abortAuthentication(c *gin.Context, err error) {
	if isAuthenticationError(err) {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			models.Response{Err: "authentication required"},
		)
		return
	}

	c.AbortWithStatusJSON(
		http.StatusInternalServerError,
		models.Response{Err: "authentication unavailable"},
	)
}

func authenticatedUserID(c *gin.Context) (uuid.UUID, error) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("authenticated user ID is missing")
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("authenticated user ID has an invalid type")
	}

	return userID, nil
}

// getOffset fetches the offset from query params and convert to int
func getOffset(c *gin.Context) (int, error) {
	var offset = 0
	var err error

	if c.Query("offset") != "" {
		offset, err = strconv.Atoi(c.Query("offset"))
		if err != nil {
			return 0, err
		}
	}
	return offset, nil
}
