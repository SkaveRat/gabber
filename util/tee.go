package util

import "io"

type Tee struct {
	R io.Reader
	W io.Writer
}

func (t Tee) Read(p []byte) (n int, err error) {
	n, err = t.R.Read(p)
	if n > 0 {
		t.W.Write(p[0:n])
		t.W.Write([]byte("\n"))
	}
	return
}
