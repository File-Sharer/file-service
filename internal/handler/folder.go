package handler

import (
	"net/http"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/gin-gonic/gin"
)

type foldersCreateReq struct {
	FolderID *string `json:"folderId"`
	Name     string  `json:"name" binding:"required"`
	Public   *bool   `json"isPublic"`
}

func (h *Handler) foldersCreate(c *gin.Context) {
	user := h.getUser(c)

	var input foldersCreateReq
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	folder, err := h.services.Folder.Create(c.Request.Context(), model.Folder{
		FolderID: input.FolderID,
		CreatorID: user.ID,
		Name: input.Name,
		Public: input.Public,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *Handler) foldersGetContents(c *gin.Context) {
	user := h.getUser(c)

	id := c.Param("id")

	contents, err := h.services.Folder.GetFolderContents(c.Request.Context(), id, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, *contents)
}

func (h *Handler) foldersGetUser(c *gin.Context) {
	user := h.getUser(c)

	folders, err := h.services.Folder.GetUserFolders(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folders)
}
