package service

type AddPermissionData struct {
	UserToken   string
	FileID      string
	UserID      string
	UserToAddID string
}

type DeletePermissionData struct {
	FileID         string
	UserID         string
	UserToDeleteID string
}
