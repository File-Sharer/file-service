package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

func (h *Handler) requestFileFromFileStorage(url string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request to file-storage: %s", err.Error())
	}
	req.Header.Set("X-Internal-Token", os.Getenv("X_INTERNAL_TOKEN"))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file(%s) from file-storage: %s", url, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file-storage server responded with status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}

type getZippedFolderReq struct {
	Path string `json:"path"`
}

func (h *Handler) requestZippedFolderFromFileStorage(path string) (io.ReadCloser, error) {
	bodyJSON, err := json.Marshal(getZippedFolderReq{Path: path})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %s", err.Error())
	}

	req, err := http.NewRequest(http.MethodGet, viper.GetString("fileStorage.origin") + "/folders", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create new request to get zipped folder by path(%s) from file-storage: %s", path, err.Error())
	}
	req.Header.Set("X-Internal-Token", os.Getenv("X_INTERNAL_TOKEN"))

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder by path(%s) from file-storage: %s", path, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("file-storage server responded with status %d: %s", resp.StatusCode, string(body))
	}

	return resp.Body, nil
}
