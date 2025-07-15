package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	fileObj.Public = isPublic
	fileObj.DownloadFilename = downloadFilename

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "file is require"})
		return
	}

	createdFile, err := h.services.File.Create(c.Request.Context(), &fileObj, file, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": createdFile})
}

func (h *Handler) filesGet(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), fileID, user.ID)
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

	fileID := c.Param("file_id")

	file, err := h.services.File.ProtectedFindByID(c.Request.Context(), fileID, user.ID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "error": err.Error()})
		return
	}

	f, err := h.getFileDownload(file.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}
	defer f.Close()

	c.Header("filename", file.DownloadFilename)
	io.Copy(c.Writer, f)
}

func (h *Handler) getFileDownload(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request to file-storage: %s", err.Error())
	}
	req.Header.Set("X-Internal-Token", os.Getenv("X_INTERNAL_TOKEN"))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from storage: %s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("storage server responded with status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
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

	fileID := c.Param("file_id")

	if err := h.services.File.Delete(c.Request.Context(), fileID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesDeletePermission(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")
	userToDeleteID := c.Param("user_id")

	if err := h.services.File.DeletePermission(c.Request.Context(), &service.DeletePermissionData{
		FileID: fileID,
		UserID: user.ID,
		UserToDeleteID: userToDeleteID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}

func (h *Handler) filesFindPermissionsToFile(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")

	permissions, err := h.services.File.FindPermissionsToFile(c.Request.Context(), fileID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil, "data": permissions})
}

func (h *Handler) filesTogglePublic(c *gin.Context) {
	user := h.getUser(c)

	fileID := c.Param("file_id")

	if err := h.services.File.TogglePublic(c.Request.Context(), fileID, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "error": nil})
}
