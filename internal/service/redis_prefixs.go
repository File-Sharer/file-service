package service

import "fmt"

var (
	filePrefix = "file:%s" // file:<fileID>
	filePermissionPrefix = "%s:%s" // <fileID>:<username>
	userFilesPrefix = "user-files:%s" // <userID>
	fileCreateDelayPrefix = "file-creating-delay-for:%s" // <userID>
	filePermissionsPrefix = "file-permissions:%s" // <fileID>
	spacePrefix = "space:%s" // <userID>
	spaceSizePrefix = "space-size:%s" // <userID>
	folderPrefix = "folder:%s" // <folderID>
	folderPermissionPrefix = "folder-permission:%s:%s" // <folderID>:<username>
	folderPermissionsPrefix = "folder-permissions:%s" // <folderID>
	folderContentsPrefix = "folder-contents:%s" // <folderID>
	userFoldersPrefix = "user-folders:%s" // <userID>
	spaceByUsernamePrefix = "space-by-username:%s" // <username>
)

func FilePrefix(fileID string) string {
	return fmt.Sprintf(filePrefix, fileID)
}

func FilePermissionPrefix(fileID, username string) string {
	return fmt.Sprintf(filePermissionPrefix, fileID, username)
}

func UserFilesPrefix(userID string) string {
	return fmt.Sprintf(userFilesPrefix, userID)
}

func FileCreateDelayPrefix(userID string) string {
	return fmt.Sprintf(fileCreateDelayPrefix, userID)
}

func FilePermissionsPrefix(fileID string) string {
	return fmt.Sprintf(filePermissionsPrefix, fileID)
}

func SpacePrefix(userID string) string {
	return fmt.Sprintf(spacePrefix, userID)
}

func SpaceSizePrefix(userID string) string {
	return fmt.Sprintf(spaceSizePrefix, userID)
}

func FolderPrefix(id string) string {
	return fmt.Sprintf(folderPrefix, id)
}

func FolderPermissionPrefix(folderID, username string) string {
	return fmt.Sprintf(folderPermissionPrefix, folderID, username)
}

func FolderPermissionsPrefix(folderID string) string {
	return fmt.Sprintf(folderPermissionsPrefix, folderID)
}

func FolderContentsPrefix(folderID string) string {
	return fmt.Sprintf(folderContentsPrefix, folderID)
}

func UserFoldersPrefix(userID string) string {
	return fmt.Sprintf(userFoldersPrefix, userID)
}

func SpaceByUsernamePrefix(username string) string {
	return fmt.Sprintf(spaceByUsernamePrefix, username)
}
