package logwriter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Open returns a new log writer.
func Open(options Options) (*Writer, error) {

	if len(options.Dir) == 0 {
		return nil, errors.New("dir must be specified")
	}

	if len(options.Ext) == 0 {
		options.Ext = logName
	}

	if options.Namer == nil {
		options.Namer = DefaultNamer()
	}

	// mkdir
	if err := os.MkdirAll(options.Dir, 0755); err != nil {
		return nil, err
	}

	w := &Writer{
		options: options,
	}

	var rotate bool

	metapath := filepath.Join(options.Dir, meatName)
	metdata, err := os.ReadFile(metapath)
	// has no meta file
	if os.IsNotExist(err) {
		rotate = true
		// error
	} else if err != nil {
		return nil, err
		// exist
	} else {
		// read metadata
		split := strings.Split(string(metdata), "\n")
		if len(split) != 3 {
			rotate = true
			goto finally
		}

		// last filename
		active := split[0]

		// written bytes
		written, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			rotate = true
			goto finally
		}

		// last write time
		ts, err := strconv.ParseInt(split[2], 10, 64)
		if err != nil {
			rotate = true
			goto finally
		}

		// open active file
		file, err := os.OpenFile(active, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
		if err != nil {
			return nil, err
		}

		w.active = file
		w.bytesWritten = written
		w.lastWrite = ts

		_ = os.Remove(metapath)
	}

finally:
	if rotate || w.hasOld(time.Now()) {
		if err := w.rotate(); err != nil {
			return nil, err
		}
	}

	return w, nil
}

// Writer is a thread-safe and append-only writer, implementing the io.Writer interface.
type Writer struct {
	options Options

	active *os.File

	bytesWritten int64
	lastWrite    int64

	mu     sync.Mutex
	closed bool
}

func (w *Writer) Write(bs []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, os.ErrClosed
	}

	now := time.Now()
	if w.hasOld(now) {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	// write log
	n, err = w.active.Write(bs)
	if err != nil {
		return 0, err
	}

	w.bytesWritten += int64(n)
	w.lastWrite = now.UnixNano()

	// decide if to sync file
	if w.bytesWritten > w.options.SyncThreshold {
		if err := w.active.Sync(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

// check if is needed to rotate, must be called in lock
func (w *Writer) hasOld(now time.Time) bool {
	// check threshold
	if w.options.MaxFileSize > 0 && w.bytesWritten >= w.options.MaxFileSize {
		return true
	}
	// check time
	lastTime := time.Unix(0, w.lastWrite)
	if w.options.Duration > 0 && now.Sub(lastTime) >= w.options.Duration {
		return true
	}
	return false
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return os.ErrClosed
	}

	// record metadata
	metaFile, err := os.OpenFile(filepath.Join(w.options.Dir, meatName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer metaFile.Close()

	if _, err := metaFile.WriteString(fmt.Sprintf("%s\n%d\n%d", w.active.Name(), w.bytesWritten, w.lastWrite)); err != nil {
		return err
	}

	if w.active != nil {
		return w.active.Close()
	}

	w.closed = true
	return nil
}

func (w *Writer) rotate() error {
	// sync and closed old active
	if w.active != nil {
		if err := w.active.Sync(); err != nil {
			return err
		}
		if err := w.active.Close(); err != nil {
			return err
		}
	}

	// apply the namer
	now := time.Now()
	name := w.options.Namer(now, w.options.Ext)
	filename := filepath.Join(w.options.Dir, name)

	// open the new active file
	newActive, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	// reset state
	w.active = newActive
	w.bytesWritten = 0
	w.lastWrite = time.Now().UnixNano()

	return nil
}
