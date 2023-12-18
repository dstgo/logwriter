package logwriter

import (
	"time"
)

const (
	meatName = "meta"
	logName  = "log"
)

func DefaultOptions(dir string) Options {
	return Options{
		Dir:      dir,
		Ext:      logName,
		Duration: time.Hour * 48,
		// 10MB
		MaxFileSize:   10 * (1 << 20),
		SyncThreshold: 0,
		Namer:         DefaultNamer(),
	}
}

type Options struct {
	Dir string
	Ext string
	// Duration determines the writer's frequency of opening a new log file, disabled if it is 0
	Duration time.Duration
	// MaxFileSize is the threshold to determine when to open a new log file, disabled if it is 0
	MaxFileSize int64
	// SyncThreshold is threshold to determine the writer will call fsync after how many bytes have been written,
	// sync always if it is 0
	SyncThreshold int64
	// Namer returns the filename when opening a new log file
	Namer func(t time.Time, ext string) string
}

func DefaultNamer() func(t time.Time, ext string) string {
	return func(t time.Time, ext string) string {
		tf := t.Format("2006_01_02_T_15_04_05")
		return tf + "." + ext
	}
}
