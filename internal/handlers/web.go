package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebHandler serves HTML pages.
type WebHandler struct{}

// Dashboard renders the main dashboard page.
func (h *WebHandler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"Title": "Dashboard"})
}

// Login renders the login page.
func (h *WebHandler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}
