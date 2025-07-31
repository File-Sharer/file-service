package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	pb "github.com/File-Sharer/file-service/hasher_pbs"
	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.Logger
	services *service.Service
	hasherClient pb.HasherClient
	httpClient *http.Client
}

func New(logger *zap.Logger, services *service.Service, hasherClient pb.HasherClient) *Handler {
	return &Handler{
		logger: logger,
		services: services,
		hasherClient: hasherClient,
		httpClient: &http.Client{},
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.SetTrustedProxies(nil)

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{viper.GetString("frontend.origin")},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		AllowMethods: []string{"POST", "GET", "PUT", "DELETE", "PATCH"},
		ExposeHeaders: []string{"filename"},
	}))

	api := router.Group("/api")
	{
		usersSpaces := api.Group("/users-spaces")
		{
			usersSpaces.PATCH("/level", h.mwSLInternal, h.usersSpacesUpdateLevel)

			usersSpaces.GET("/level", h.mwAuth, h.usersSpacesGetLevel)
		}

		folders := api.Group("/folders")
		folders.Use(h.mwAuth)
		{
			folders.POST("", h.foldersCreate)
			folders.GET("/:id/contents", h.foldersGetContents)
			folders.GET("", h.foldersGetUser)
			folders.GET("/:id/permissions", h.foldersGetPermissions)
			folders.PUT("/:id/:username", h.foldersAddPermission)
			folders.DELETE("/:id/:username", h.foldersDeletePermission)
			folders.GET("/:id/dl", h.foldersGetZipped)
		}

		files := api.Group("/files")
		files.Use(h.mwAuth)
		{
			files.POST("", h.filesCreate)
			files.GET("/:file_id", h.filesGet)
			files.GET("", h.filesFindUser)
			files.GET("/:file_id/dl", h.filesDownload)
			files.PUT("/:file_id/:username", h.filesAddPermission)
			files.DELETE("/:file_id", h.filesDelete)
			files.DELETE("/:file_id/:username", h.filesDeletePermission)
			files.GET("/:file_id/permissions", h.filesFindPermissionsToFile)
			files.PATCH("/:file_id/togglepub", h.filesTogglePublic)
		}
	}

	return router
}

func (h *Handler) getToken(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("no provided token")
	}

	token := strings.Split(header, " ")[1]
	if token == "" {
		return "", errors.New("no provided token")
	}

	return token, nil
}

func (h *Handler) getUserDataFromToken(ctx context.Context, token string) (*model.UserSpace, string, error) {
	decoded, err := h.hasherClient.DecodeJWT(context.Background(), &pb.DecodeJWTReq{Secret: os.Getenv("HASHER_SECRET"), Jwt: token})
	if err != nil {
		return nil, "", err
	}

	userSpace, err := h.services.UserSpace.Get(ctx, decoded.UserId)
	if err != nil {
		return nil, "", err
	}

	return userSpace, decoded.Role, nil
}

func (h *Handler) getUserSpace(c *gin.Context) *model.UserSpace {
	userSpaceReq, ok := c.Get("user-space")
	if !ok {
		return nil
	}

	userSpace, ok := userSpaceReq.(model.UserSpace)
	if !ok {
		return nil
	}

	return &userSpace
}

func (h *Handler) getUserRole(c *gin.Context) *string {
	userRoleReq, ok := c.Get("user-role")
	if !ok {
		return nil
	}

	userRole, ok := userRoleReq.(string)
	if !ok {
		return nil
	}

	return &userRole
}
