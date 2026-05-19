package controllers

import (
	"betterov2/models"
	"log"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
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

// getOffset fetches the offset from query params and convert to int
func getOffset(c *gin.Context) (int, error) {
	var offset int = 0
	var err error

	if c.Query("offset") != "" {
		offset, err = strconv.Atoi(c.Query("offset"))
		if err != nil {
			return 0, err
		}
	}
	return offset, nil
}

// round rounds a float64 to 2 decimal places
func round(x float64) float64 {
	return math.Round(x*100) / 100
}
