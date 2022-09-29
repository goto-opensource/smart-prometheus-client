package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap_AddGet(t *testing.T) {
	theMap := newTagMap()

	theMap.Add([]string{"abc", "toto"}, "value1")
	value, present := theMap.Get([]string{"abc", "toto"})
	assert.True(t, present)
	assert.Equal(t, "value1", value)

	value, present = theMap.Get([]string{"abcd", "toto"})
	assert.False(t, present)
	assert.Equal(t, "", value)
}

func TestMap_AddDelete(t *testing.T) {
	theMap := newTagMap()

	theMap.Add([]string{"abc", "toto"}, "value1")
	theMap.Add([]string{"abcd", "toto1"}, "value2")
	theMap.Add([]string{"abce", "toto2"}, "value3")

	value, present := theMap.Get([]string{"abcd", "toto1"})
	assert.True(t, present)
	assert.Equal(t, "value2", value)

	theMap.Delete([]string{"abcd", "toto1"})

	value, present = theMap.Get([]string{"abcd", "toto1"})
	assert.False(t, present)
	assert.Equal(t, "", value)

	value, present = theMap.Get([]string{"abc", "toto"})
	assert.True(t, present)
	assert.Equal(t, "value1", value)

	value, present = theMap.Get([]string{"abce", "toto2"})
	assert.True(t, present)
	assert.Equal(t, "value3", value)
}

func TestMap_AddGetCollision(t *testing.T) {
	theMap := newTagMap()
	theMap.hash = func(s []string) uint64 { return 1 }

	theMap.Add([]string{"abc", "toto"}, "value1")
	value, present := theMap.Get([]string{"abc", "toto"})
	assert.True(t, present)
	assert.Equal(t, "value1", value)

	value, present = theMap.Get([]string{"abcd", "toto"})
	assert.False(t, present)
	assert.Equal(t, "", value)
}

func TestMap_AddDeleteCollision(t *testing.T) {
	theMap := newTagMap()
	theMap.hash = func(s []string) uint64 { return 1 }

	theMap.Add([]string{"abcd", "toto1"}, "value2")

	value, present := theMap.Get([]string{"abcd", "toto1"})
	assert.True(t, present)
	assert.Equal(t, "value2", value)

	theMap.Delete([]string{"abcd", "toto1"})

	value, present = theMap.Get([]string{"abcd", "toto1"})
	assert.False(t, present)
	assert.Equal(t, "", value)

	theMap.Add([]string{"abce", "toto2"}, "value3")
	theMap.Add([]string{"abc", "toto"}, "value1")
	theMap.Add([]string{"abcd", "toto1"}, "value2")

	value, present = theMap.Get([]string{"abcd", "toto1"})
	assert.True(t, present)
	assert.Equal(t, "value2", value)

	theMap.Delete([]string{"abcd", "toto1"})

	value, present = theMap.Get([]string{"abc", "toto"})
	assert.True(t, present)
	assert.Equal(t, "value1", value)

	value, present = theMap.Get([]string{"abce", "toto2"})
	assert.True(t, present)
	assert.Equal(t, "value3", value)
}
