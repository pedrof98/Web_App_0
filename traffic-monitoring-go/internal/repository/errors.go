package repository

import (
	"errors"
	"time"

)

// Common repository errors
var (
	ErrNotFound = errors.New("entity not found")
	ErrDuplicate = errors.New("entity already exists")
	ErrForeignKey = errors.New("foreign key constraint violation")
)


// timeFromTimestamp converts an int64 timestamp to a time.Time
func timeFromTimestamp(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0)
}

