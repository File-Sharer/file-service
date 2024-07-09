package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/File-Sharer/file-service/internal/model"
	"github.com/File-Sharer/file-service/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Handler struct {
	services *service.Service
}

func New(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.SetTrustedProxies(nil)

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{viper.GetString("frontend.origin")},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		AllowMethods: []string{"POST", "GET", "PATCH", "DELETE"},
		ExposeHeaders: []string{"filename"},
	}))

	api := router.Group("/api")
	{
		files := api.Group("/files")
		{
			files.POST("", h.mwAuth, h.filesCreate)
			files.GET("/:id", h.mwAuth, h.filesGet)
			files.GET("", h.mwAuth, h.filesFindUser)
			files.GET("/:id/dl", h.mwAuth, h.filesDownload)
			files.PATCH("/:file_id/:user_id", h.mwAuth, h.filesAddPermission)
			files.DELETE("/:id", h.mwAuth, h.filesDelete)
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
	host := viper.GetString("userService.host")
	endpoint := "/api/user"

	client := &http.Client{}

	req := &http.Request{
		Proto: "HTTP/1.1",
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host: host,
			Path: endpoint,
		},
		Header: map[string][]string{
			"Authorization": {"Bearer " + token},
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user data: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user data: status code %d", res.StatusCode)
	}

	var userRes model.UserRes
	if err := json.NewDecoder(res.Body).Decode(&userRes); err != nil {
		return nil, fmt.Errorf("failed to decode user data: %w", err)
	}

	if !userRes.Ok {
		return nil, fmt.Errorf("failed to get user data: %s", userRes.Error)
	}

	return &userRes.Data, nil
}

func (h *Handler) getUser(c *gin.Context) *model.User {
	userReq, _ := c.Get("user")

	user, ok := userReq.(model.User)
	if !ok {
		return nil
	}

	return &user
}
