package common_handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ErrorResponse defines the standard error response structure.
// swagger:response ErrorResponse
type ErrorResponse struct {
	// Error message
	// example: Invalid input data
	Error string `json:"error"`
	// Optional details about the error
	// example: Field 'email' is required
	Details string `json:"details,omitempty"`
}

// getPaginationParams extracts page and limit from query parameters.
func GetPaginationParams(c *gin.Context) (page, limit int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	var err error
	page, err = strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	// Optional: Add max limit check
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

// getFilteringParams extracts filtering parameters from query parameters.
func GetFilteringParams(c *gin.Context) map[string]interface{} {
	filters := make(map[string]interface{})
	log := c.MustGet("logger").(*logrus.Logger) // Assuming logger is set in middleware

	if minAgeStr := c.Query("min_age"); minAgeStr != "" {
		minAge, err := strconv.Atoi(minAgeStr)
		if err == nil && minAge > 0 {
			filters["min_age"] = minAge
		} else {
			log.Warnf("Invalid min_age parameter: %s", minAgeStr)
		}
	}

	if maxAgeStr := c.Query("max_age"); maxAgeStr != "" {
		maxAge, err := strconv.Atoi(maxAgeStr)
		if err == nil && maxAge > 0 {
			filters["max_age"] = maxAge
		} else {
			log.Warnf("Invalid max_age parameter: %s", maxAgeStr)
		}
	}

	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}

	// TODO: добавить больше фильтров

	return filters
}
