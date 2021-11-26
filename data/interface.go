package data

import "bytes"

// CountTags defines what a TagCount handler should be able to do
type CountTags interface {
	Serialize() ([]*bytes.Buffer, error)
	ParseTags(attributes []string)
	GetTags() []string
	Len() int
}
