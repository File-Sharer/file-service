package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) fileCreate(c *gin.Context) {
	user := h.getUser(c)

	var fileObj model.File

	isPublicForm := c.PostForm("isPublic")
	isPublic, err := strconv.ParseBool(isPublicForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "isPublic is required"})
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

	if err := h.services.File.Create(c.Request.Context(), &fileObj, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) fileGet(c *gin.Context) {
	user := h.getUser(c)

	id := c.Param("id")

	file, err := h.services.File.FindByID(c.Request.Context(), id, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": file})
}

func (h *Handler) fileDownload(c *gin.Context) {
	user := h.getUser(c)

	id := c.Param("id")

	file, err := h.services.File.FindByID(c.Request.Context(), id, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": err.Error()})
		return
	}

	filePath := filepath.Join("files/", file.Filename)

	f, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "file not found"})
		return
	}
	defer f.Close()

	c.Header("filename", file.DownloadFilename)
	io.Copy(c.Writer, f)
}

func (h *Handler) fileAddPermission(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")
	userToAdd := c.Param("user_id")
	
	if err := h.services.File.AddPermission(c.Request.Context(), fileID, user.ID, userToAdd); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}
