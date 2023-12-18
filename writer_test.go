package logwriter

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOpen_New(t *testing.T) {
	testdir := filepath.Join(os.TempDir(), "logw")
	defer func() {
		err := os.RemoveAll(testdir)
		assert.Nil(t, err)
	}()

	writer, err := Open(DefaultOptions(testdir))
	assert.Nil(t, err)
	assert.NotNil(t, writer)
	err = writer.Close()
	assert.Nil(t, err)
}

func TestOpen_Exist(t *testing.T) {
	testdir := filepath.Join(os.TempDir(), "logw")
	defer func() {
		err := os.RemoveAll(testdir)
		assert.Nil(t, err)
	}()

	writer, err := Open(DefaultOptions(testdir))
	assert.Nil(t, err)
	assert.NotNil(t, writer)
	err = writer.Close()
	assert.Nil(t, err)

	writer1, err := Open(DefaultOptions(testdir))
	assert.Nil(t, err)
	assert.NotNil(t, writer1)
	err = writer1.Close()
	assert.Nil(t, err)

	assert.EqualValues(t, writer.active.Name(), writer1.active.Name())
}

func TestWriter_Write_1(t *testing.T) {
	testdir := filepath.Join(os.TempDir(), "logw")
	writer, err := Open(DefaultOptions(testdir))
	assert.Nil(t, err)
	assert.NotNil(t, writer)

	defer func() {
		err = writer.Close()
		assert.Nil(t, err)
		err := os.RemoveAll(testdir)
		assert.Nil(t, err)
	}()

	n, err := writer.Write([]byte("abc"))
	assert.Nil(t, err)
	assert.Greater(t, n, 0)
}

func TestWriter_Write_2(t *testing.T) {
	testdir := filepath.Join(os.TempDir(), "logw")
	options := DefaultOptions(testdir)
	options.MaxFileSize = 10
	writer, err := Open(options)
	assert.Nil(t, err)
	assert.NotNil(t, writer)

	defer func() {
		err = writer.Close()
		assert.Nil(t, err)
		err := os.RemoveAll(testdir)
		assert.Nil(t, err)
	}()

	oldActive := writer.active
	n, err := writer.Write([]byte(strings.Repeat("a", 11)))
	assert.Nil(t, err)
	assert.Greater(t, n, 0)

	n1, err := writer.Write([]byte(strings.Repeat("a", 1)))
	assert.Nil(t, err)
	assert.Greater(t, n1, 0)

	newActive := writer.active

	assert.NotEqual(t, oldActive.Name(), newActive.Name())
}

func TestWriter_Write_3(t *testing.T) {
	testdir := filepath.Join(os.TempDir(), "logw")
	options := DefaultOptions(testdir)
	options.Duration = time.Second * 5
	writer, err := Open(options)
	assert.Nil(t, err)
	assert.NotNil(t, writer)

	defer func() {
		err = writer.Close()
		assert.Nil(t, err)
		err := os.RemoveAll(testdir)
		assert.Nil(t, err)
	}()

	oldActive := writer.active
	time.Sleep(time.Second * 6)

	n1, err := writer.Write([]byte(strings.Repeat("a", 1)))
	assert.Nil(t, err)
	assert.Greater(t, n1, 0)

	newActive := writer.active

	assert.NotEqual(t, oldActive.Name(), newActive.Name())
}
