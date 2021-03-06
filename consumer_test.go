package funnel

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestRollover(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	f, err := os.Open("testdata/file_84lines")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	c.Start(f)

	// testing results
	files := readTestDir(t, dir)

	sep := []byte{'\n'}
	if len(files) != 3 {
		t.Errorf("Incorrect no. of files created. Expected 3, Got %d", len(files))
	}
	for i, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		numLines := bytes.Count(data, sep)
		// First 2 files will have 40 lines
		// last one will have 4 lines
		if i < 2 {
			if numLines != 40 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 40, Got %d", i, numLines)
			}
		} else {
			if numLines != 4 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 4, Got %d", i, numLines)
			}
		}
	}
}

func TestHugeLine(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	// This file also contains arabic, indian and tibetan characters
	// to test any ascii-utf8 codec incompatibility
	targetBytes, err := ioutil.ReadFile("testdata/file_bigline")
	if err != nil {
		t.Fatal(err)
		return
	}
	r := bytes.NewReader(targetBytes)
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		// removing the newline character at the end
		cmp := bytes.Compare(data, targetBytes)
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", string(targetBytes), string(data))
		}
	}
}

func TestNewLines(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	// This file also contains arabic, indian and tibetan characters
	// to test any ascii-utf8 codec incompatibility
	targetBytes, err := ioutil.ReadFile("testdata/file_newline")
	if err != nil {
		t.Fatal(err)
		return
	}
	r := bytes.NewReader(targetBytes)
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		cmp := bytes.Compare(data, targetBytes)
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", string(targetBytes), string(data))
		}
	}
}

func TestEmptyString(t *testing.T) {
	dir, c := setupTest(t)
	defer os.RemoveAll(dir)

	r := strings.NewReader("")
	c.Start(r)

	// testing results
	files := readTestDir(t, dir)
	if len(files) != 1 {
		t.Errorf("Incorrect no. of files created. Expected 1, Got %d", len(files))
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		cmp := bytes.Compare(data, []byte(""))
		if cmp != 0 {
			t.Errorf("Incorrect string found. Expected- %s, Found- %s", "", string(data))
		}
	}
}

func TestRolloverSerial(t *testing.T) {
	dir, c := setupTest(t)
	c.Config.FileRenamePolicy = "serial"
	defer os.RemoveAll(dir)

	f, err := os.Open("testdata/file_84lines")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	c.Start(f)

	// testing results
	files := readTestDir(t, dir)

	sep := []byte{'\n'}
	if len(files) != 3 {
		t.Errorf("Incorrect no. of files created. Expected 3, Got %d", len(files))
	}
	for i, file := range files {
		data, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			t.Fatal(err)
			continue
		}
		numLines := bytes.Count(data, sep)
		// First file will have 4 lines
		// rest will have 40 lines
		if i == 0 {
			if numLines != 4 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 4, Got %d", i, numLines)
			}
		} else {
			if numLines != 40 {
				t.Errorf("Incorrect no. of lines created in file #%d. Expected 40, Got %d", i, numLines)
			}
		}
	}
}

func TestSendInterruptSerial(t *testing.T) {
	// TODO
}

// Internal helper functions
func setupTest(t *testing.T) (string, *Consumer) {
	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
		return "", nil
	}

	c := &Consumer{
		Config: &Config{
			DirName:                  dir,
			ActiveFileName:           "out.log",
			RotationMaxLines:         40,
			RotationMaxBytes:         1000000,
			FlushingTimeIntervalSecs: 5,
			FileRenamePolicy:         "timestamp",
			MaxAge:                   int64(1 * 60 * 60),
			MaxCount:                 500,
		},
		LineProcessor: &NoProcessor{},
	}
	return dir, c
}

func readTestDir(t *testing.T, dir string) []os.FileInfo {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
		return nil
	}
	return files
}
