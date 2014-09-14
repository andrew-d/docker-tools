package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseContainerImage(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	err = parseContainerMapImage(&q, "foo")
	assert.NoError(t, err)
	assert.Equal(t, q.Image, "foo")

	err = parseContainerMapImage(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerDependencies(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	input := []interface{}{
		"foo",
		"bar:baz",
	}

	err = parseContainerMapDependencies(&q, input)
	assert.NoError(t, err)
	assert.Equal(t, q.Dependencies, []DepConfig{
		{"foo", "foo"},
		{"bar", "baz"},
	})

	input = []interface{}{
		"foo:bar:baz",
	}
	err = parseContainerMapDependencies(&q, input)
	assert.EqualError(t, err, "Unknown format for dependency entry 0")

	input = []interface{}{
		1234,
	}
	err = parseContainerMapDependencies(&q, input)
	assert.EqualError(t, err, "Unknown value type in array: int")

	err = parseContainerMapDependencies(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerEnv(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	input := []interface{}{
		"FOO=BAR",
		"ABC=123=345",
	}

	err = parseContainerMapEnv(&q, input)
	assert.NoError(t, err)
	assert.Equal(t, q.Env, []EnvConfig{
		{"FOO", "BAR"},
		{"ABC", "123=345"},
	})

	input = []interface{}{
		"FOOBAR",
	}
	err = parseContainerMapEnv(&q, input)
	assert.EqualError(t, err, "Env entry 0 not in form KEY=VAL")

	input = []interface{}{
		1234,
	}
	err = parseContainerMapEnv(&q, input)
	assert.EqualError(t, err, "Unknown value type in array: int")

	err = parseContainerMapEnv(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerPorts(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	input := []interface{}{
		12345,
		"4444",
		"123:456",
		"ipaddr:999:888",
	}

	err = parseContainerMapPorts(&q, input)
	assert.NoError(t, err)
	assert.Equal(t, q.Ports, []PortConfig{
		{"0.0.0.0", 12345, 12345},
		{"0.0.0.0", 4444, 4444},
		{"0.0.0.0", 123, 456},
		{"ipaddr", 999, 888},
	})

	err = parseContainerMapPorts(&q, []interface{}{999999999})
	assert.EqualError(t, err, "Port 0 out of range: 999999999")

	err = parseContainerMapPorts(&q, []interface{}{"999999999"})
	assert.Error(t, err)

	err = parseContainerMapPorts(&q, []interface{}{"asdf"})
	assert.Error(t, err)

	err = parseContainerMapPorts(&q, []interface{}{"asdf:zzzz"})
	assert.Error(t, err)

	err = parseContainerMapPorts(&q, []interface{}{"asdf:1234:zzzz"})
	assert.Error(t, err)

	err = parseContainerMapPorts(&q, []interface{}{"123:456:789:000"})
	assert.EqualError(t, err, "Unknown port format for port 0")

	var invalid []string
	err = parseContainerMapPorts(&q, []interface{}{invalid})
	assert.EqualError(t, err, "Unknown value type in array: []string")

	err = parseContainerMapPorts(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerMount(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	input := []interface{}{
		"/a/b/c:/foo/bar",
		"/d/e/f:/baz/123:ro",
		"/quux:/other:rw",
	}

	err = parseContainerMapMount(&q, input)
	assert.NoError(t, err)
	assert.Equal(t, q.Mount, []MountConfig{
		{"/a/b/c", "/foo/bar", MountTypeReadWrite},
		{"/d/e/f", "/baz/123", MountTypeReadOnly},
		{"/quux", "/other", MountTypeReadWrite},
	})

	input = []interface{}{"/foo:/bar:rr"}
	err = parseContainerMapMount(&q, input)
	assert.EqualError(t, err, "Mount entry 0 has invalid mount type: rr")

	input = []interface{}{"badformat"}
	err = parseContainerMapMount(&q, input)
	assert.EqualError(t, err, "Mount entry 0 not in form /host/dir:/container/dir[:type]")

	input = []interface{}{1234}
	err = parseContainerMapMount(&q, input)
	assert.EqualError(t, err, "Unknown value type in array: int")

	err = parseContainerMapMount(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerMountFrom(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	input := []interface{}{
		"abc",
		"123",
	}

	err = parseContainerMapMountFrom(&q, input)
	assert.NoError(t, err)
	assert.Equal(t, q.MountFrom, []string{
		"abc",
		"123",
	})

	input = []interface{}{1234}
	err = parseContainerMapMountFrom(&q, input)
	assert.EqualError(t, err, "Unknown value type in array: int")

	err = parseContainerMapMountFrom(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerPrivileged(t *testing.T) {
	t.Parallel()

	var q Container
	var err error

	var input interface{} = true

	err = parseContainerMapPrivileged(&q, input)
	assert.NoError(t, err)
	assert.Exactly(t, q.Privileged, true)

	input = false
	err = parseContainerMapPrivileged(&q, input)
	assert.NoError(t, err)
	assert.Exactly(t, q.Privileged, false)

	err = parseContainerMapPrivileged(&q, 1234)
	assert.EqualError(t, err, "Unknown value type: int")
}

func TestParseContainerMap(t *testing.T) {
	t.Parallel()

	var err error

	input := map[interface{}]interface{}{
		"image":        "",
		"dependencies": []interface{}{},
		"env":          []interface{}{},
		"ports":        []interface{}{},
		"mount":        []interface{}{},
		"mount-from":   []interface{}{},
		"privileged":   false,
	}

	_, err = parseContainerMap("test", input)
	assert.NoError(t, err)

	input = map[interface{}]interface{}{
		1234: "asdf",
	}
	_, err = parseContainerMap("test", input)
	assert.EqualError(t, err, "Unknown key in config for container test: 1234")

	input = map[interface{}]interface{}{
		"otherkey": "",
	}
	_, err = parseContainerMap("test", input)
	assert.EqualError(t, err, "Unknown key in config for container test: otherkey")
}

func TestParseContainer(t *testing.T) {
	t.Parallel()

	var err error
	var c *Container

	c, err = parseContainer("foo", "theimage")
	assert.NoError(t, err)
	assert.Equal(t, c, &Container{
		Name:  "foo",
		Image: "theimage",
	})

	input := map[interface{}]interface{}{
		"image": "img123",
	}
	c, err = parseContainer("foo", input)
	assert.NoError(t, err)
	assert.Equal(t, c, &Container{
		Name:  "foo",
		Image: "img123",
	})

	c, err = parseContainer("bad", 1234)
	assert.EqualError(t, err, "Unknown type for 'containers' key: int")
}
