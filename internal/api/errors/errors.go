/*
uniwish.com/internal/api/errors

contains domain errors
*/
package errors

import "errors"

var ErrUnavailable = errors.New("service unavailable")
var ErrStoreUnsupported = errors.New("store unsupported")
var ErrInputInvalid = errors.New("input invalid")
var ErrScrapeFailed = errors.New("Scrape failed")
