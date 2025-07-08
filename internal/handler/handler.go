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
)

type Handler struct {
	services *service.Service
	hasherClient pb.HasherClient
	httpClient *http.Client
}

func New(services *service.Service, hasherClient pb.HasherClient) *Handler {
	return &Handler{
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
		AllowMethods: []string{"POST", "GET", "PUT", "DELETE"},
		ExposeHeaders: []string{"filename"},
	}))

	api := router.Group("/api")
	{
		files := api.Group("/files")
		files.Use(h.mwAuth)
		{
			files.POST("", h.filesCreate)
			files.GET("/:file_id", h.filesGet)
			files.GET("", h.filesFindUser)
			files.GET("/:file_id/dl", h.filesDownload)
			files.PUT("/:file_id/:user_id", h.filesAddPermission)
			files.DELETE("/:file_id", h.filesDelete)
			files.DELETE("/:file_id/permission", h.filesDeletePermission)
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

func (h *Handler) getUserDataFromToken(token string) (*model.User, error) {
	user, err := h.hasherClient.DecodeJWT(context.Background(), &pb.DecodeJWTReq{Secret: os.Getenv("HASHER_SECRET"), Jwt: token})
	if err != nil {
		return nil, err
	}

	return &model.User{ID: user.UserId, Role: user.Role}, nil
}

func (h *Handler) getUser(c *gin.Context) *model.User {
	userReq, _ := c.Get("user")

	user, ok := userReq.(model.User)
	if !ok {
		return nil
	}

	return &user
}
