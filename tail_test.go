package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func createTmpFile(t *testing.T) *os.File {
	f, err := ioutil.TempFile(os.TempDir(), "test_go_tail_")
	if err != nil {
		t.Fatal("failed creating temp file:", err)
	}
	return f
}

func TestIsDone(t *testing.T) {
	tail := &Tail{
		done: make(chan struct{}),
	}
	if tail.IsClosed() {
		t.Fatal("unexpected tail isDone")
	}
	close(tail.done)
	if !tail.IsClosed() {
		t.Fatal("expected tail isDone")
	}
}

func TestOutputLine(t *testing.T) {
	tail := &Tail{
		Line: make(chan string, 1),
		done: make(chan struct{}),
	}
	out := tail.outputLine([]byte("line"))
	if !out {
		t.Fatal("expected true value")
	}
	line, ok := <-tail.Line
	if !ok || line != "line" {
		t.Fatal("unexpected result")
	}

	var done bool
	go func(t *testing.T) {
		// first output should succeed, but will fill channel
		tail.outputLine([]byte("line1"))
		// second line output should block on full Line channel
		// It will be unblocked by closing done channel and out should be false
		out = tail.outputLine([]byte("line2"))
		done = true
	}(t)

	time.Sleep(250 * time.Millisecond)
	close(tail.done)
	time.Sleep(250 * time.Millisecond)
	if done != true && out != false {
		t.Fatal("unexpected outputLine termination state")
	}
}

func TestOpenFile(t *testing.T) {
	tail := &Tail{fileName: "non_existing_test_file"}
	err := tail.openFile()
	if !os.IsNotExist(err) {
		t.Fatal("unexpected error:", err)
	}
	if tail.file != nil {
		t.Fatal("expected nil tail.file")
	}

	testFile := createTmpFile(t)
	fileName := testFile.Name()
	testFile.Write([]byte("line"))
	testFile.Close()
	defer os.Remove(fileName)

	tail.fileName = fileName
	err = tail.openFile()
	if err != nil {
		t.Fatal("expected nil error")
	}
	if tail.lastSize != 4 {
		t.Fatal("expected file size", 4, "got", tail.lastSize)
	}
	if tail.file == nil {
		t.Fatal("expected non-nil tail.file")
	}
	tail.file.Close()
	tail.file = nil

	testError = errors.New("test error")
	err = tail.openFile()
	if err != testError {
		t.Fatal("expected error ", testError, "got", err)
	}
	tail.file.Close()
	tail.file = nil
	testError = nil
}

func TestScanLines(t *testing.T) {
	testFile := createTmpFile(t)
	fileName := testFile.Name()
	testFile.Write([]byte("line 1\nline 2\r\nline 3\nline 4"))
	testFile.Close()
	defer os.Remove(fileName)
	lines := []string{"line 1", "line 2", "line 3", "line 4"}

	savedBufInitSize := bufInitSize
	bufInitSize = 3
	tail := NewTail(fileName)

	for _, l := range lines[:3] {
		line, ok := <-tail.Line
		if !ok || line != l {
			t.Fatal("expected", l, "got", line)
		}
	}
	tail.Close()
	time.Sleep(500 * time.Millisecond)

	savedLineChanSize := lineChanSize
	lineChanSize = 2
	tail = NewTail(fileName)
	time.Sleep(500 * time.Millisecond)
	tail.Close()
	time.Sleep(500 * time.Millisecond)

	bufInitSize = savedBufInitSize
	lineChanSize = savedLineChanSize
}

func TestReadLines(t *testing.T) {

}

/*
func TestTailWithNonExistingFile(t *testing.T) {
	fileName := "toto_xyz"
	tail := NewTail(fileName)
	_, ok := <-tail.Line
	if ok {
		t.Fatal("expect Line channel to be closed")
	}

	err, ok := <-tail.Error
	if !ok {
		t.Fatal("expected to read an error from Error channel")
	}
	if err == nil {
		t.Fatal("unexpected nil error")
	}
	if !os.IsNotExist(err) {
		t.Fatal("unexpected error:", err)
	}
}

func TestTailCloseWithEmptyFile(t *testing.T) {
	testFile := createTmpFile(t)
	fileName := testFile.Name()
	defer testFile.Close()
	defer os.Remove(fileName)
	tail := NewTail(fileName)
	tail.Close()
	_, ok := <-tail.Line
	if ok {
		t.Fatal("unexpected line read from file")
	}
}

func TestTailCloseWithNonEmptyFile(t *testing.T) {
	testFile := createTmpFile(t)
	fileName := testFile.Name()
	//testFile.Write([]byte("line 1\nline 2\r\nline 3\nline 4"))
	defer testFile.Close()

	tail := NewTail(fileName)
	tail.Close()
	_, ok := <-tail.Line
	if ok {
		t.Fatal("unexpected line read from file")
	}

	testFile.Write([]byte("line 1\nline 2\r\nline 3\nline 4"))
	tail = NewTail(fileName)

}
*/

// func readAllLines(tail *Tail) []string {
// 	lines := make([]string, 0, 10)
// 	for {
// 		select {
// 		case <-tail.done:
// 			return lines
// 		case line := <-tail.Line:
// 			lines = append(lines, line)
// 		}
// 	}
// }

// func TestScanLines(t *testing.T) {
// 	testFile := createTmpFile(t)
// 	fileName := testFile.Name()
// 	testFile.Write([]byte("line 1\nline 2\r\nline 3\nline 4"))
// 	testFile.Close()

// 	var err error
// 	tail := NewTail(fileName)
// 	tail.file, err = os.Open(testFile.Name())
// 	if err != nil {
// 		t.Fatal("failed opening test file:", err)
// 	}

// 	err = tail.scanLines()
// 	if err != io.EOF {
// 		t.Fatal("unexpected error:", err)
// 	}
// 	lines := make([]string, 0, 10)

// }
