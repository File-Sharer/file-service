package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Space Level Internal Middleware
func (h *Handler) mwSLInternal(c *gin.Context) {
	header := strings.TrimSpace(c.GetHeader("X_Internal_Token"))
	if !strings.HasPrefix(header, "SL ") {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": errNoToken.Error()})
		c.Abort()
		return
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": errNoToken.Error()})
		c.Abort()
		return
	}

	token := parts[1]
	if token != os.Getenv("SL_Internal_Token") {
		c.JSON(http.StatusForbidden, gin.H{"ok": false})
		c.Abort()
		return
	}

	c.Next()
}
