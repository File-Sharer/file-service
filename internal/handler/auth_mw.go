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

	user, err := h.getUserDataFromToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		c.Abort()
		return
	}

	c.Set("user", *user)

	c.Next()
}
