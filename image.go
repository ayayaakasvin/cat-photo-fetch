package catphotofetch

import (
	"bytes"
	"io"
)

type Image struct {
	ContentType string
	Data        []byte
}

func (i *Image) ReaderCloser() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(i.Data))
}
