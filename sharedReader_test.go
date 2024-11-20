package efincore

import (
	"io"
	"strings"
	"sync"
	"testing"
)

func TestRBody(t *testing.T) {
	body := io.NopCloser(strings.NewReader("test-val"))
	reader := newRBody(body)

	var client1Got string
	var client1RestGot string
	var client2Got string
	var client2RestGot string

	syncChan := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		bytes := make([]byte, 4)
		if _, err := reader.Read(bytes); err != nil {
			t.Fatalf("read 1 from reader1 failed: %v", err)
		}

		client1Got = string(bytes)
		syncChan <- struct{}{}

		<-syncChan
		if _, err := reader.Read(bytes); err != nil {
			t.Fatalf("read 2 from reader1 failed: %v", err)
		}
		client1RestGot = string(bytes)
	}()

	reader2 := reader.Clone()
	go func() {
		defer wg.Done()

		<-syncChan
		bytes := make([]byte, 4)
		if _, err := reader2.Read(bytes); err != nil {
			t.Fatalf("read 1 from reader2 failed: %v", err)
		}

		client2Got = string(bytes)

		if _, err := reader2.Read(bytes); err != nil {
			t.Fatalf("read 2 from reader2 failed: %v", err)
		}

		client2RestGot = string(bytes)
		syncChan <- struct{}{}
	}()

	wg.Wait()

	expected := "test"
	expectedRest := "-val"

	if client1Got != expected {
		t.Errorf("client1: got '%s', expected '%s'", client1Got, expected)
	}

	if client1RestGot != expectedRest {
		t.Errorf("client1 rest: got '%s', expected '%s'", client1RestGot, expectedRest)
	}

	if client2Got != expected {
		t.Errorf("client2: got '%s', expected '%s'", client2Got, expected)
	}

	if client2RestGot != expectedRest {
		t.Errorf("client2 rest: got '%s', expected '%s'", client2RestGot, expectedRest)
	}
}
