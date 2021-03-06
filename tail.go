package main

// Tail return lines from a text file into a string channel and
// keep returning lines when appended at runtime.

import (
	"io"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	lineChanSize           = 100
	bufInitSize            = 2048
	testError1, testError2 error
)

type Tail struct {
	fileName string            // name of file to read lines from
	file     *os.File          // file we read lines from
	Line     chan string       // channel to return line strings
	Error    chan error        // channel to report errors
	done     chan struct{}     // channel to signal Tail close
	buf      []byte            // reading buffer
	nbytes   int               // number valid bytes in buffer
	lastSize int64             // last file files
	watcher  *fsnotify.Watcher // watcher on file
}

func NewTail(fileName string) *Tail {
	t := &Tail{
		fileName: fileName,
		Line:     make(chan string, lineChanSize),
		Error:    make(chan error, 1),
		done:     make(chan struct{}),
		buf:      make([]byte, bufInitSize),
	}
	go runTail(t)
	return t
}

// Close terminates monitoring the file and close the channel. Has no
// effect if the Tail is already closed.
func (t *Tail) Close() {
	if !t.IsClosed() {
		close(t.done)
		if t.watcher != nil {
			t.watcher.Close()
			t.watcher = nil
		}
	}
}

// IsClosed return true when the channel has been closed.
func (t *Tail) IsClosed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

// outputLine return true if successfully output b as a line, and false
// if tail has been closed.
func (t *Tail) outputLine(b []byte) bool {
	select {
	case <-t.done:
		return false
	case t.Line <- string(b):
		return true
	}
}

func (t *Tail) openFile() error {
	var err error
	t.file, err = os.Open(t.fileName)
	if err != nil {
		return err
	}
	stat, err := t.file.Stat()
	if err != nil || testError1 != nil {
		if err == nil {
			err = testError1
		}
		return err
	}
	t.lastSize = stat.Size()
	return nil
}

// readLines output lines read from the file until an error or io.EOF is met,
// or done is closed. It returns the error, or nil when done is closed.
func (t *Tail) readLines() error {
	for {
		if len(t.buf) == t.nbytes {
			tmp := make([]byte, len(t.buf)*2)
			copy(tmp, t.buf)
			t.buf = tmp
		}
		nbytes := t.nbytes
		n, err := t.file.Read(t.buf[t.nbytes:])
		t.nbytes += n
		if err != nil {
			return err
		}
		buf := t.buf[:t.nbytes]
		begPos := 0
		for i := nbytes; i < t.nbytes; i++ {
			var line []byte
			if buf[i] == '\n' {
				if i > 0 && buf[i-1] == '\r' {
					line = buf[begPos : i-1]
				} else {
					line = buf[begPos:i]
				}
				if !t.outputLine(line) {
					return nil
				}
				begPos = i + 1
			}
		}
		if begPos != 0 {
			t.nbytes = copy(t.buf, buf[begPos:])
		}
	}
}

// runTail is the goroutine that read lines from the file.
func runTail(t *Tail) {
	var err error
	defer func() {
		if t.file != nil {
			t.file.Close()
			t.file = nil
		}
		if err != nil {
			t.Error <- err
		}
		t.Close()
	}()

	// try starting watcher
	if t.watcher, err = fsnotify.NewWatcher(); err != nil || testError1 != nil {
		if err == nil {
			err = testError1
		}
		return
	}

	// try open file
	if err = t.openFile(); err != nil {
		return
	}
	// read all existing lines in file
	if err = t.readLines(); err != io.EOF {
		return
	}

	// start watching file to detect appending or file renaming
	if err = t.watcher.Add(t.fileName); err != nil || testError2 != nil {
		if err == nil {
			err = testError2
		}
		return
	}

	// loop over file change events
	for !t.IsClosed() {
		var event fsnotify.Event
		var ok bool

		select {
		case <-t.done:
			return
		case event, ok = <-t.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// text was appended to file
				if err = t.readLines(); err != io.EOF {
					return
				}
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				// a file rotation occured
				if t.nbytes > 0 && !t.outputLine(t.buf[:t.nbytes]) {
					return
				}
				t.nbytes = 0
				t.file.Close()
				t.watcher.Remove(t.fileName)
				delay, maxDelay := time.Second, 30*time.Second
				for delay < maxDelay {
					if err = t.openFile(); err == nil {
						break
					}
					time.Sleep(delay)
					delay *= 2
				}
				if err != nil {
					return
				}
				t.watcher.Add(t.fileName)
				if err = t.readLines(); err != io.EOF {
					return
				}
			}
		case err, ok = <-t.watcher.Errors:
			return
		}
	}
}
