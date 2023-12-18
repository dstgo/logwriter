# logwriter
An implementation of Writer that can split based on file size and time duration

## install
```bash
go get -u github.com/dstgo/logwriter@latest
```

## usage

a simple usage that is used with `log/slog` package
```go
package main

import (
	"github.com/dstgo/logwriter"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func main() {
	writer, err := logwriter.Open(logwriter.DefaultOptions(filepath.Join(os.TempDir(), "lw")))
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	var w io.WriteCloser
	w = writer

	handler := slog.NewTextHandler(w, nil)
	logger := slog.New(handler)
	logger.Info("logging")
}
```