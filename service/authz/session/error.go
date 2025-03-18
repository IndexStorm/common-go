package session

import "errors"

var ErrSessionNotFound = errors.New("session was not found")
var ErrSessionExpired = errors.New("session has been expired")
