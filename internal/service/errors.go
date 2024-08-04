package service

import "errors"

var (
	errNoAccess = errors.New("you have no access")
	errFileIsTooBig = errors.New("file is too big, max upload size: 256MB")
	errUserNotFound = errors.New("user not found")
	errMaxUploadsReached = errors.New("you have reached your max uploads size: 10KB")
)
