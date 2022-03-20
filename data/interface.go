package data

// CountTags defines what a TagCount handler should be able to do
type CountTags interface {
	Serialize(buffSlice *BufferSlice, index string) error
	ParseTags(attributes []string)
	GetTags() []string
	Len() int
}
