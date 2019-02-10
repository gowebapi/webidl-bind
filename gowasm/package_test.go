package gowasm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortPackageName(t *testing.T) {
	assert.Equal(t, "dom", shortPackageName("dom"))
	assert.Equal(t, "dom", shortPackageName("github.com/webapi/dom"))
}

func TestSourceFilename(t *testing.T) {
	src := &Source{
		name:    "hello.go",
		Package: "github.com/gowebapi/webapi",
	}
	filename, inc := src.Filename("")
	assert.True(t, inc)
	assert.Equal(t, "github.com/gowebapi/webapi/hello.go", filename)

	filename, inc = src.Filename("foo")
	assert.False(t, inc)

	filename, inc = src.Filename("github.com/gowebapi")
	assert.True(t, inc)
	assert.Equal(t, "webapi/hello.go", filename)

	filename, inc = src.Filename("github.com/gowebapi/")
	assert.True(t, inc)
	assert.Equal(t, "webapi/hello.go", filename)

	filename, inc = src.Filename("github.com/gowebapi/webapi")
	assert.True(t, inc)
	assert.Equal(t, "hello.go", filename)

	filename, inc = src.Filename("github.com/gowebapi/webapi/")
	assert.True(t, inc)
	assert.Equal(t, "hello.go", filename)
}
