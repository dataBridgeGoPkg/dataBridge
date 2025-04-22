package utils

import "strconv"

// ParseID converts a string ID to int64, returns 0 if invalid
func ParseID(id string) int64 {
	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0
	}
	return parsedID
}
