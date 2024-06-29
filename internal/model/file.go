package model

import "time"

type File struct {
	ID        string    `json:"id"`
	CreatorID string    `json:"creatorId"`
	IsPublic  bool      `json:"isPublic"`
	DateAdded time.Time `json:"dateAdded"`
}
