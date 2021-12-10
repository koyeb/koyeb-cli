package idmapper2

import "regexp"

const (
	// UUIDv4 is the regular expressions for an UUID v4.
	UUIDv4 string = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)

var (
	// RxUUIDv4 is a compiled regular expression for an UUID v4.
	RxUUIDv4 = regexp.MustCompile(UUIDv4)
)

// IsUUIDv4 checks if the string is a UUID version 4.
func IsUUIDv4(val string) bool {
	return RxUUIDv4.MatchString(val)
}
