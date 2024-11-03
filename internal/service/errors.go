package service

import "errors"

var (
	errNoAccess = errors.New("you have no access")
	errFileIsTooBig = errors.New("file is too big, max upload size: 256MB")
	errUserNotFound = errors.New("user not found")
	errMaxUploadsReached = errors.New("you have reached your max uploads size: 10KB")
	errWaitDelay = errors.New("please wait until the timeout is over, it is 2 mins for creating files")
	errCantAddPermissionForYourself = errors.New("you cannot add permission to your file for yourself")
)
