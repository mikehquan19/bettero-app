package controllers

import (
	"betterov2/models"
	"log"
	"strconv"
	"time"

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

// getSummaryDates parses the summary dates from query params
func getSummaryDates(c *gin.Context) (models.SummaryDates, error) {
	var dates models.SummaryDates
	var err error

	for _, field := range []struct {
		dest *time.Time
		name string
	}{
		{dest: &dates.CurrStart, name: "curr_start"},
		{dest: &dates.CurrEnd, name: "curr_end"},
		{dest: &dates.PrevStart, name: "prev_start"},
		{dest: &dates.PrevEnd, name: "prev_end"},
	} {
		*field.dest, err = time.Parse("2006-01-02", c.Query(field.name))
		if err != nil {
			return models.SummaryDates{}, err
		}
	}

	return dates, nil
}
