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
