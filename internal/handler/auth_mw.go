package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) mwAuth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "no provided token"})
		c.Abort()
		return
	}

	token := strings.Split(header, " ")[1]
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "no provided token"})
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
