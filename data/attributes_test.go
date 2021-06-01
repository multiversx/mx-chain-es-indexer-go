package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareAttributesShouldWork(t *testing.T) {
	t.Parallel()

	attributes := []byte("tags:test,free,fun;description:This is a test description for an awesome nft")
	prepared := NewAttributesDTO(attributes)
	require.Equal(t, &Attributes{
		"tags":        []string{"test", "free", "fun"},
		"description": []string{"This is a test description for an awesome nft"},
	}, prepared)

	attributes = []byte("tags:test,free,fun;description: ")
	prepared = NewAttributesDTO(attributes)
	require.Equal(t, &Attributes{
		"tags":        []string{"test", "free", "fun"},
		"description": []string{" "},
	}, prepared)

	attributes = []byte("description:;")
	prepared = NewAttributesDTO(attributes)
	require.Equal(t, &Attributes{
		"description": []string{""},
	}, prepared)

	attributes = []byte("")
	prepared = NewAttributesDTO(attributes)
	require.Nil(t, prepared)

	attributes = []byte(";")
	prepared = NewAttributesDTO(attributes)
	require.Nil(t, prepared)

	attributes = []byte("  ")
	prepared = NewAttributesDTO(attributes)
	require.Equal(t, &Attributes{
		"  ": []string{"  "},
	}, prepared)

	attributes = []byte("attribute")
	prepared = NewAttributesDTO(attributes)
	require.Equal(t, &Attributes{
		"attribute": []string{"attribute"},
	}, prepared)
}
