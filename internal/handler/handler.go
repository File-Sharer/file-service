package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

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
		AllowAllOrigins: true,
		AllowHeaders: []string{"Authorization", "Content-Type"},
		ExposeHeaders: []string{"filename"},
	}))

	api := router.Group("/api")
	{
		file := api.Group("/file")
		{
			file.POST("", h.mwAuth, h.fileCreate)
			file.GET("/:id", h.mwAuth, h.fileGet)
			file.GET("/dl/:id", h.mwAuth, h.fileDownload)
			file.PATCH("/:file_id/:user_id", h.mwAuth, h.fileAddPermission)
		}
	}

	return router
}

func (h *Handler) getUserDataFromToken(token string) (*model.User, error) {
	target := viper.GetString("userService.host")
	endpoint := "/api/user"

	client := &http.Client{}

	req := &http.Request{
		Proto: "HTTP/1.1",
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host: target,
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
