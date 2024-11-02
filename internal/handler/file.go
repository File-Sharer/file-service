package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *Handler) filesCreate(c *gin.Context) {
	user := h.getUser(c)

	var fileObj model.File

	isPublicForm := c.PostForm("isPublic")
	isPublic, err := strconv.ParseBool(isPublicForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "isPublic option is required"})
		return
	}

	downloadFilename := strings.TrimSpace(c.PostForm("downloadFilename"))
	if downloadFilename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "download filename is required"})
		return
	}

	fileObj.CreatorID = user.ID
	fileObj.IsPublic = isPublic
	fileObj.DownloadFilename = downloadFilename

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "file is require"})
		return
	}

	createdFile, err := h.services.File.Create(c.Request.Context(), &fileObj, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": createdFile})
}

func (h *Handler) filesGet(c *gin.Context) {
	user := h.getUser(c)

	id := c.Param("id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), id, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": file})
}

func (h *Handler) filesFindUser(c *gin.Context) {
	user := h.getUser(c)

	files, err := h.services.File.FindUserFiles(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": files})
}

func (h *Handler) filesDownload(c *gin.Context) {
	user := h.getUser(c)

	id := c.Param("id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), id, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": err.Error()})
		return
	}

	filePath := filepath.Join("files/" + file.CreatorID, file.Filename)

	f, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "file not found"})
		return
	}
	defer f.Close()

	c.Header("filename", file.DownloadFilename)
	io.Copy(c.Writer, f)
}

func (h *Handler) filesAddPermission(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")
	userToAddID := c.Param("user_id")

	userToken, err := h.getToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "error": err.Error()})
		return
	}
	
	data := &service.AddPermissionData{
		UserToken: userToken,
		FileID: fileID,
		UserID: user.ID,
		UserToAddID: userToAddID,
	}
	if err := h.services.File.AddPermission(c.Request.Context(), data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesDelete(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("id")

	if err := h.services.File.Delete(c.Request.Context(), fileID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}
