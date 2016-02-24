package bongo

import (
	"reflect"
	"strings"
	"unicode"
)

// GetBsonName returms bson name
func GetBsonName(field reflect.StructField) string {
	tag := field.Tag.Get("bson")
	tags := strings.Split(tag, ",")

	if len(tags[0]) > 0 {
		return tags[0]
	}
	return lowerInitial(field.Name)
}

// lowerInitial returns lower cases first char of string
func lowerInitial(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}
