package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type reqUsersSpacesUpdateLevel struct {
	UserID string `json:"userId" binding:"required"`
	Level  uint8  `json:"level" binding:"required,min=1,max=3"`
}

func (h *Handler) usersSpacesUpdateLevel(c *gin.Context) {
	var input reqUsersSpacesUpdateLevel
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	if err := h.services.UserSpace.UpdateLevel(c.Request.Context(), input.UserID, input.Level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}
