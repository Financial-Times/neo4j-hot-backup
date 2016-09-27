package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestore(t *testing.T) {
	readDirContents(t, "./test-data")

	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	n := newMockNeoBackup()

	restoreErr := restoreDirectory(n, "backup-test", dir)
	assert.NoError(t, restoreErr)

	extractedDir := filepath.Join(dir, "backup-test")

	inFile, _ := os.Open(filepath.Join(extractedDir, "SHA256SUMS"))
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := scanner.Text()
		split := strings.Split(text, " ")
		checkSha(t, extractedDir, split[2], split[0])
	}
}

func checkSha(t *testing.T, dir string, fileName string, sha string) {
	fmt.Printf("File: '%s' expected SHA: '%s'\n", fileName, sha)
	path := filepath.Join(dir, fileName)
	fb, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	assert.NotEmpty(t, fb)
	hasher := sha256.New()
	hasher.Write(fb)
	actualSha := hex.EncodeToString(hasher.Sum(nil))
	assert.Equal(t, sha, actualSha, "File: '%s' expected SHA: '%s'\n", fileName, sha)
}

func readDirContents(t *testing.T, dir string) {
	fmt.Printf("Dir: %s\n", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
		t.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

type fakeNeoBackup struct {
	s3bucket string
	s3dir    string
}

func newMockNeoBackup() *fakeNeoBackup {
	return &fakeNeoBackup{"s3-bucket", "s3-dir"}
}

func (n *fakeNeoBackup) readS3(path string) (string, io.ReadCloser, http.Header, error) {
	fb, err := ioutil.ReadFile("./test-data/backup-test.tar.sz")
	reader := ioutil.NopCloser(bytes.NewBuffer(fb))
	return "path", reader, nil, err
}
