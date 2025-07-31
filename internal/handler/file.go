package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) filesCreate(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	var fileObj model.File

	folderID := strings.TrimSpace(c.PostForm("folderId"))
	fileObj.FolderID = new(string)
	if folderID != "" {
		*fileObj.FolderID = folderID
	} else {
		fileObj.FolderID = nil
	}

	isPublicForm := c.PostForm("isPublic")
	isPublic, err := strconv.ParseBool(isPublicForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "isPublic option type must be boolean"})
		return
	}

	downloadName := strings.TrimSpace(c.PostForm("downloadName"))
	if downloadName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "download filename is required"})
		return
	}

	fileObj.CreatorID = userSpace.UserID
	fileObj.Public = &isPublic
	fileObj.DownloadName = downloadName

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "file is required"})
		return
	}

	createdFile, err := h.services.File.Create(c.Request.Context(), fileObj, file, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": createdFile})
}

func (h *Handler) filesGet(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	fileID := c.Param("file_id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), fileID, *userRole, *userSpace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": file})
}

func (h *Handler) filesFindUser(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	files, err := h.services.File.FindUserFiles(c.Request.Context(), userSpace.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": files})
}

func (h *Handler) filesDownload(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	fileID := c.Param("file_id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), fileID, *userRole, *userSpace)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": err.Error()})
		return
	}

	f, err := h.requestFileFromFileStorage(file.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	defer f.Close()

	c.Header("filename", file.DownloadName)
	io.Copy(c.Writer, f)
}

func (h *Handler) filesAddPermission(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	fileID := c.Param("file_id")
	userToAddName := c.Param("username")
	
	data := service.AddPermissionData{
		ResourceID: fileID,
		UserSpace: *userSpace,
		UserToAddName: userToAddName,
	}
	if err := h.services.File.AddPermission(c.Request.Context(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesDelete(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	fileID := c.Param("file_id")

	if err := h.services.File.Delete(c.Request.Context(), fileID, *userRole, *userSpace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesDeletePermission(c *gin.Context) {
	userSpace := h.getUserSpace(c)
	userRole := h.getUserRole(c)

	fileID := c.Param("file_id")
	userToDeleteName := c.Param("username")

	if err := h.services.File.DeletePermission(c.Request.Context(), service.DeletePermissionData{
		ResourceID: fileID,
		UserID: userSpace.UserID,
		UserRole: *userRole,
		UserToDeleteName: userToDeleteName,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesFindPermissionsToFile(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	fileID := c.Param("file_id")

	permissions, err := h.services.File.FindPermissionsToFile(c.Request.Context(), fileID, userSpace.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, permissions)
}

func (h *Handler) filesTogglePublic(c *gin.Context) {
	userSpace := h.getUserSpace(c)

	fileID := c.Param("file_id")

	if err := h.services.File.TogglePublic(c.Request.Context(), fileID, userSpace.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}
