package service

import "errors"

var (
	errFileNotFound = errors.New("file not found")
	errInternal = errors.New("internal server error")
	errNoAccess = errors.New("you have no access")
	errFileIsTooBig = errors.New("file is too big, max upload size: 256MB")
	errUserNotFound = errors.New("user not found")
	errWaitDelay = errors.New("please wait until the timeout is over, it is 2 mins for creating files")
	errCantAddPermissionForYourself = errors.New("you cannot add permission to your file for yourself")
	errFailedToUploadFileToFileStorage = errors.New("failed to upload file to file storage")
	errYouDoNotHaveEnoughSpace = errors.New("you don not have enough space")
)
