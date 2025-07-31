package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type foldersCreateReq struct {
	FolderID *string `json:"folderId"`
	Name     string  `json:"name" binding:"required"`
	Public   bool   `json"isPublic"`
}

func (h *Handler) foldersCreate(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	var input foldersCreateReq
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	folder, err := h.services.Folder.Create(c.Request.Context(), model.Folder{
		FolderID: input.FolderID,
		CreatorID: userSpace.UserID,
		Name: input.Name,
		Public: &input.Public,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *Handler) foldersGetContents(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	id := c.Param("id")

	contents, err := h.services.Folder.GetFolderContents(c.Request.Context(), id, *userRole, *userSpace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, *contents)
}

func (h *Handler) foldersGetUser(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	folders, err := h.services.Folder.GetUserFolders(c.Request.Context(), userSpace.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, folders)
}

func (h *Handler) foldersGetPermissions(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	folderID := c.Param("id")

	permissions, err := h.services.Folder.GetPermissions(c.Request.Context(), folderID, userSpace.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (h *Handler) foldersAddPermission(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	folderID := c.Param("id")
	userToAddName := c.Param("username")

	if err := h.services.Folder.AddPermission(c.Request.Context(), service.AddPermissionData{
		ResourceID: folderID,
		UserSpace: *userSpace,
		UserRole: *userRole,
		UserToAddName: userToAddName,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) foldersDeletePermission(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	folderID := c.Param("id")
	userToDeleteName := c.Param("username")

	if err := h.services.Folder.DeletePermission(c.Request.Context(), service.DeletePermissionData{
		ResourceID: folderID,
		UserID: userSpace.UserID,
		UserRole: *userRole,
		UserToDeleteName: userToDeleteName,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) foldersGetZipped(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	folderID := c.Param("id")

	folder, err := h.services.Folder.ProtectedFindByID(c.Request.Context(), folderID, *userRole, *userSpace)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": err.Error()})
		return
	}

	parts := strings.Split(folder.URL, viper.GetString("fileStorage.origin") + "/files/")
	if len(parts) != 2 {
		h.logger.Sugar().Errorf("invalid folder(%s) URL: %v", folder.ID, parts)
		return
	}
	path := parts[1]

	f, err := h.requestZippedFolderFromFileStorage(path)
	if err != nil {
		h.logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": errInternal})
		return
	}
	defer f.Close()

	c.Header("filename", folder.Name)
	io.Copy(c.Writer, f)
}
