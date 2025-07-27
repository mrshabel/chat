package util

import "bytes"

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// SanitizeWSMessage trims whitespaces from the raw websocket message and replaces all occurrences of new lines with space
func SanitizeWSMessage(msg []byte) string {
	msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
	return string(msg)
}
