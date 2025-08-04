package service

import "github.com/File-Sharer/file-service/internal/model"

type AddPermissionData struct {
	ResourceID    string
	UserSpace     model.FullUserSpace
	UserRole      string
	UserToAddName string
}

type DeletePermissionData struct {
	ResourceID       string
	UserID           string
	UserRole         string
	UserToDeleteName string
}
