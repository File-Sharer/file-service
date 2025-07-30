package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) mwAuth(c *gin.Context) {
	token, err := h.getToken(c)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": err.Error()})
		c.Abort()
		return
	}

	userSpace, userRole, err := h.getUserDataFromToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		c.Abort()
		return
	}

	c.Set("user-space", *userSpace)
	c.Set("user-role", userRole)

	c.Next()
}
