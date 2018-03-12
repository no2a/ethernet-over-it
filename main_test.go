package main

import (
	"bytes"
	"github.com/songgao/water"
	"io"
	"testing"
)

var (
	ifce *water.Interface
)

func ensureTAP() *water.Interface {
	if ifce == nil {
		ifce = startTAP("test", false)
	}
	return ifce
}

func BenchmarkSendToTAP(b *testing.B) {
	ifce := ensureTAP()
	C := make(chan []byte)
	go func() {
		bytes := make([]byte, 14)
		for i := 0; i < b.N; i++ {
			C <- bytes
		}
		close(C)
	}()
	sendToTAP(ifce, C)
}

type (
	RepeatReader struct {
		io.Reader
		Buffer *bytes.Reader
		Count  *int
	}
)

func InitRepeatReader(b []byte, count int) RepeatReader {
	r := RepeatReader{}
	r.Buffer = bytes.NewReader(b)
	r.Count = new(int)
	*r.Count = count
	return r
}

func (r RepeatReader) Read(p []byte) (int, error) {
	if *r.Count <= 0 {
		return 0, io.EOF
	}
	n, err := r.Buffer.Read(p)
	if err == io.EOF {
		*r.Count--
		if *r.Count > 0 {
			r.Buffer.Seek(0, 0)
			if n == 0 {
				n, err = r.Buffer.Read(p)
				if err == io.EOF {
					err = nil
				}
			}
		}
	}
	return n, err
}

func TestRepeatReader(t *testing.T) {
	src := make([]byte, 20)
	r := InitRepeatReader(src, 2)

	b := make([]byte, 100)
	n, err := r.Read(b)
	t.Logf("%d %s", n, err)
	if n != 20 || err != nil {
		t.Fail()
	}

	n, err = r.Read(b)
	t.Logf("%d %s", n, err)
	if n != 20 || err != nil {
		t.Fail()
	}

	n, err = r.Read(b)
	t.Logf("%d %s", n, err)
	if n != 0 || err != io.EOF {
		t.Fail()
	}
}

func _BenchmarkRecv(b *testing.B, size int, size2 int) {
	totsize := 2 + size + 2 + size2
	src := make([]byte, totsize)
	src[0] = byte(size % 256)
	src[1] = byte((size / 256) % 256)
	if size2 == 0 {
		src = src[:2+size]
	} else {
		src[2+size] = byte(size2 % 256)
		src[2+size+1] = byte((size2 / 256) % 256)
	}
	r := InitRepeatReader(src, b.N)

	c := make(chan []byte)

	go func() {
		for _ = range c {
		}
	}()

	recv(r, c)
}

func BenchmarkRecv14(b *testing.B) {
	_BenchmarkRecv(b, 14, 0)
}

func BenchmarkRecv1500(b *testing.B) {
	_BenchmarkRecv(b, 1500, 0)
}

func BenchmarkRecv9000(b *testing.B) {
	_BenchmarkRecv(b, 9000, 0)
}

func BenchmarkRecv14_14(b *testing.B) {
	_BenchmarkRecv(b, 1500, 1000)
}

func BenchmarkRecv1500_1000(b *testing.B) {
	_BenchmarkRecv(b, 1500, 1000)
}

func BenchmarkLenBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lenBytes(1500)
	}
}

func BenchmarkBytesLen(b *testing.B) {
	bytes := make([]byte, 2)
	bytes[0] = 123
	bytes[1] = 7
	for i := 0; i < b.N; i++ {
		bytesLen(bytes)
	}
}
