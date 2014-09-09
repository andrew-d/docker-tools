package main

import (
	"bytes"
	"io"
)

// LineStreamer applies prefixes/postfixes to lines before writing them to
// an underlying writer.
type LineStreamer struct {
	out     io.Writer
	prefix  string
	postfix string
}

// NewLineStreamer creates a new LineStreamer
func NewLineStreamer(out io.Writer, prefix, postfix string) *LineStreamer {
	ret := &LineStreamer{
		out:     out,
		prefix:  prefix,
		postfix: postfix,
	}
	return ret
}

func (l *LineStreamer) Write(p []byte) (n int, err error) {
	var buf bytes.Buffer
	var line, outLine string

	if n, err = buf.Write(p); err != nil {
		return
	}

	for {
		line, err = buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return
		}

		outLine = l.prefix + line + l.postfix
		_, err = io.WriteString(l.out, outLine)
		if err != nil {
			return
		}
	}

	return
}

var _ io.Writer = &LineStreamer{}
