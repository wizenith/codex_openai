package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebHandler serves HTML pages.
type WebHandler struct{}

// Dashboard renders the main dashboard page.
func (h *WebHandler) Dashboard(c *gin.Context) {
	// Check if it's an HTMX request for stats
	if c.GetHeader("HX-Request") != "" && c.GetHeader("HX-Target") == "stats" {
		// This would normally get stats from the database
		// For now, return the stats partial
		c.HTML(http.StatusOK, "partials/stats.html", gin.H{
			"Total":      0,
			"Pending":    0,
			"Processing": 0,
			"Completed":  0,
			"Failed":     0,
		})
		return
	}
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"Title": "Dashboard"})
}

// Login renders the login page.
func (h *WebHandler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}
