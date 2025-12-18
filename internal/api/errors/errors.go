/*
uniwish.com/internal/api/errors

contains domain api errors
*/
package errors

import "errors"

var ErrUnavailable = errors.New("service unavailable")
var ErrStoreUnsupported = errors.New("store unsupported")
var ErrInputInvalid = errors.New("input invalid")
