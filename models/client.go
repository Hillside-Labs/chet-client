package models

import "strings"

const (
	ClientNameSep = ":"
)

func NewClientName(user *User, name string, description string) string {
	return strings.Join([]string{
		"chet",
		user.Email,
		name,
	}, ClientNameSep)
}
