package model

import "time"

type File struct {
	ID                string    `json:"id"`
	MainFolderID      *string   `json:"mainFolderId"`
	FolderID          *string   `json:"folderId"`
	CreatorID         string    `json:"creatorId"`
	Size              int64     `json:"size"`
	URL               string    `json:"url"`
	Public            *bool     `json:"public"`
	Filename          string    `json:"filename"`
	DownloadFilename  *string   `json:"downloadFilename"`
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
