package proio

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteLZ4IterateFile(t *testing.T) {
	writeIterateFile(LZ4, t)
}

func TestWriteGZIPIterateFile(t *testing.T) {
	writeIterateFile(GZIP, t)
}

func TestWriteUncompIterateFile(t *testing.T) {
	writeIterateFile(UNCOMPRESSED, t)
}

func writeIterateFile(comp Compression, t *testing.T) {
	nEvents := 5

	tmpDir, err := ioutil.TempDir("", "proiotest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "writeIterateFile")

	writer, err := Create(tmpFile)
	if err != nil {
		t.Error(err)
	}
	writer.SetCompression(comp)
	event := NewEvent()
	for i := 0; i < nEvents; i++ {
		writer.Push(event)
	}
	writer.Close()

	nEvents = 0

	reader, err := Open(tmpFile)
	if err != nil {
		t.Error(err)
	}
	for range reader.ScanEvents() {
		nEvents++
	}
	reader.Close()

	if nEvents != 5 {
		t.Errorf("nEvents is %v instead of 5", nEvents)
	}
}

func TestCreateFileInEmptyDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "proiotest")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpDir = filepath.Join(tmpDir, "nonExistant")
	tmpFile := filepath.Join(tmpDir, "nonExistant")

	writer, err := Create(tmpFile)
	if err == nil {
		t.Errorf("No error thrown for creating file in non-existent directory.  Path is \"%v\"", tmpFile)
		writer.Close()
	}
}
