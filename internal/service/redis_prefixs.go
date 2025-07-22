package service

import "fmt"

var (
	filePrefix = "file:%s" // file:<fileID>
	permissionPrefix = "%s:%s" // <fileID>:<userID>
	userFilesPrefix = "user-files:%s" // user-files:<userID>
	fileCreateDelayPrefix = "file-creating-delay-for:%s" // file-creating-delay-for:<userID>
	filePermissionsPrefix = "permissions-to:%s" // permissions-to:<fileID>
	spacePrefix = "space:%s" // <userID>
	spaceSizePrefix = "space-size:%s" // <userID>
)

func FilePrefix(fileID string) string {
	return fmt.Sprintf(filePrefix, fileID)
}

func PermissionPrefix(fileID string, userID string) string {
	return fmt.Sprintf(permissionPrefix, fileID, userID)
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
