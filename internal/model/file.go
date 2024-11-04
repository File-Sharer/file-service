package model

import "time"

type File struct {
	ID                string    `json:"id"`
	CreatorID         string    `json:"creatorId"`
	IsPublic          bool      `json:"isPublic"`
	Filename          string    `json:"filename"`
	DownloadFilename  string    `json:"downloadFilename"`
	DateAdded         time.Time `json:"dateAdded"`
}

type DeleteFileReq struct {
	FileID   string `json:"id"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

type Permission struct {
	FileID string `json:"fileId"`
	UserID string `json:"userId"`
}
