package testutils

import "time"

func Int64Pointer(f int64) *int64 {
	return &f
}

func Float64Pointer(f float64) *float64 {
	return &f
}

func BoolPointer(b bool) *bool {
	return &b
}

func TimePointer(b time.Time) *time.Time {
	return &b
}
