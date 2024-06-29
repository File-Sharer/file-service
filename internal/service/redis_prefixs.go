package service

import "fmt"

var (
	filePrefix = "file:" // fileID
	permissionPrefix = "%s:%s" // fileID:userID
)

func PermissionPrefix(fileID string, userID string) string {
	return fmt.Sprintf(permissionPrefix, fileID, userID)
}
