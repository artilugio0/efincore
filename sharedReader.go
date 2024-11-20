package efincore

import (
	"io"
	"sync"
)

type RBody struct {
	inner io.ReadCloser
	bytes *[]byte
	done  *bool
	mutex *sync.Mutex
	index int
}

func newRBody(r io.ReadCloser) *RBody {
	done := false
	return &RBody{
		inner: r,
		mutex: &sync.Mutex{},
		done:  &done,
		bytes: &[]byte{},
	}
}

func (rb *RBody) Read(p []byte) (int, error) {
	// WARN: using read in a multithreaded environment is probably a bad idea
	// better make the clients use GetBytes

	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	if !*rb.done {
		if err := rb.readall(); err != nil {
			return 0, err
		}
	}

	bytesLen := len(*rb.bytes)
	targetLen := len(p)

	if rb.index >= bytesLen {
		return 0, io.EOF
	}

	var i int
	for i = 0; i < targetLen && i+rb.index < bytesLen; i++ {
		p[i] = (*rb.bytes)[rb.index+i]
	}

	rb.index += i

	return i, nil
}

func (rb *RBody) GetBytes() ([]byte, error) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	if !*rb.done {
		if err := rb.readall(); err != nil {
			return nil, err
		}
	}

	return *rb.bytes, nil
}

func (rb *RBody) readall() error {
	// WARN: this should be called only with the mutex locked

	b, err := io.ReadAll(rb.inner)
	if err != nil {
		return err
	}

	*rb.bytes = b
	*rb.done = true

	// TODO: check if this is necessary
	rb.inner.Close()

	return nil
}

func (rb *RBody) Rewind() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.index = 0
}

func (rb *RBody) Close() error {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.inner.Close()
	rb.index = 0

	return nil
}

func (rb *RBody) Clone() *RBody {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	return &RBody{
		inner: rb.inner,
		bytes: rb.bytes,
		done:  rb.done,
		mutex: rb.mutex,
		index: 0,
	}
}
