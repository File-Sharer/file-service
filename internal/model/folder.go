package model

import "time"

type Folder struct {
	ID           string    `json:"id"`
	MainFolderID *string   `json:"mainFolderId"`
	FolderID     *string   `json:"folderId"`
	CreatorID    string    `json:"creatorId"`
	URL          string    `json:"url"`
	Name         string    `json:"name"`
	Public       *bool     `json:"public"`
	CreatedAt    time.Time `json:"createdAt"`
}

type FolderContents struct {
	Files   []*File   `json:"files"`
	Folders []*Folder `json:"folders"`
}
