package idmapper

import "strings"

func getKey(id string) Key {
	return Key(getFlatID(id))
}

func getShortID(id string, length int) string {
	fkey := getFlatID(id)
	return fkey[:length]
}

func getFlatID(id string) string {
	return strings.ReplaceAll(id, "-", "")
}
