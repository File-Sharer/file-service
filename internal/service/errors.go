package service

import "errors"

var (
	errNoAccess = errors.New("you have no access")
	errFileIsTooBig = errors.New("file is too big, max upload size: 256MB")
)
