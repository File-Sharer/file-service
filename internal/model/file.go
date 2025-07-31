package model

import "time"

type File struct {
	ID           string    `json:"id"`
	MainFolderID *string   `json:"mainFolderId"`
	FolderID     *string   `json:"folderId"`
	CreatorID    string    `json:"creatorId"`
	CreatorName  *string   `json:"creatorName"`
	Size         int64     `json:"size"`
	URL          string    `json:"url"`
	Public       *bool     `json:"public"`
	Filename     *string   `json:"filename"`
	DownloadName string    `json:"downloadName"`
	DateAdded    time.Time `json:"dateAdded"`
}
