package bucket

import (
	"bytes"
	"mime/multipart"
)

type InMemoryFile struct {
	*bytes.Reader
}

func (imf *InMemoryFile) Close() error {
	return nil
}

var _ multipart.File = (*InMemoryFile)(nil)
