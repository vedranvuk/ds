// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package fs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()
	fsys := New()
	if fsys == nil {
		t.Errorf("New() returned nil")
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func TestFSOpen(t *testing.T) {
	t.Parallel()
	fsys := New()
	_, err := fsys.Create("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	file, err := fsys.Open("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	_, err = fsys.Open("nonexistent.txt")
	if err == nil {
		t.Errorf("Open() should have returned an error for nonexistent file")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func BenchmarkFSOpen(b *testing.B) {
	fsys := New()
	_, err := fsys.Create("myfile.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, err := fsys.Open("myfile.txt")
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
	}
}

func TestFSStat(t *testing.T) {
	t.Parallel()
	fsys := New()
	_, err := fsys.Create("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	fileInfo, err := fsys.Stat("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Name() != "myfile.txt" {
		t.Errorf("expected myfile.txt, got %s", fileInfo.Name())
	}

	_, err = fsys.Stat("nonexistent.txt")
	if err == nil {
		t.Errorf("Stat() should have returned an error for nonexistent file")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func BenchmarkFSStat(b *testing.B) {
	fsys := New()
	_, err := fsys.Create("myfile.txt")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fsys.Stat("myfile.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFSReadFile(t *testing.T) {
	t.Parallel()
	fsys := New()
	err := fsys.WriteFile("myfile.txt", []byte("Hello, world!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	data, err := fsys.ReadFile("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hello, world!" {
		t.Errorf("expected Hello, world!, got %s", string(data))
	}

	_, err = fsys.ReadFile("nonexistent.txt")
	if err == nil {
		t.Errorf("ReadFile() should have returned an error for nonexistent file")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func BenchmarkFSReadFile(b *testing.B) {
	fsys := New()
	err := fsys.WriteFile("myfile.txt", []byte(strings.Repeat("A", 1024)), 0644)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fsys.ReadFile("myfile.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFSReadDir(t *testing.T) {
	t.Parallel()
	fsys := New()
	fsys.Mkdir("mydir", 0755)
	fsys.WriteFile("mydir/myfile.txt", []byte(""), 0644)
	entries, err := fsys.ReadDir("mydir")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name() != "myfile.txt" {
		t.Errorf("expected myfile.txt, got %s", entries[0].Name())
	}

	entries, err = fsys.ReadDir("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}

	fsys.WriteFile("anotherfile.txt", []byte(""), 0644)
	entries, err = fsys.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range entries {
		if entry.Name() == "anotherfile.txt" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find anotherfile.txt in root dir")
	}
}

func BenchmarkFSReadDir(b *testing.B) {
	fsys := New()
	fsys.Mkdir("mydir", 0755)
	for i := 0; i < 100; i++ {
		fsys.WriteFile("mydir/myfile"+string(rune(i))+".txt", []byte(""), 0644)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fsys.ReadDir("mydir")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestDirEntry(t *testing.T) {
	t.Parallel()
	de := &dirEntry{name: "myfile.txt", isDir: false}
	if de.Name() != "myfile.txt" {
		t.Errorf("expected myfile.txt, got %s", de.Name())
	}
	if de.IsDir() {
		t.Errorf("expected false, got true")
	}
	if de.Type() != 0 {
		t.Errorf("expected 0, got %v", de.Type())
	}

	_, err := de.Info()
	if err != nil {
		t.Fatal(err)
	}

	de = &dirEntry{name: "mydir", isDir: true}
	if !de.IsDir() {
		t.Errorf("expected true, got false")
	}
	if de.Type() != fs.ModeDir {
		t.Errorf("expected fs.ModeDir, got %v", de.Type())
	}
}

func BenchmarkDirEntry(b *testing.B) {
	de := &dirEntry{name: "myfile.txt", isDir: false}
	for i := 0; i < b.N; i++ {
		de.Name()
		de.IsDir()
		de.Type()
		de.Info()
	}
}

func TestFSCreate(t *testing.T) {
	t.Parallel()
	fsys := New()
	file, err := fsys.Create("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	if file == nil {
		t.Errorf("Create() returned nil")
	}
	if _, err := fsys.Create("myfile.txt"); err != ErrExist {
		t.Errorf("expected ErrExist, got %v", err)
	}
}

func BenchmarkFSCreate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fsys := New()
		_, err := fsys.Create("myfile.txt")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFSMkdir(t *testing.T) {
	t.Parallel()
	fsys := New()
	err := fsys.Mkdir("mydir", 0755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fsys.Stat("mydir"); err != nil {
		t.Fatal(err)
	}
	if _, err := fsys.Create("mydir"); err == nil {
		t.Errorf("expected error, got nil")
	}
	if err := fsys.Mkdir("mydir", 0755); err != ErrExist {
		t.Errorf("expected ErrExist, got %v", err)
	}
}

func BenchmarkFSMkdir(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fsys := New()
		err := fsys.Mkdir("mydir", 0755)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFSWriteFile(t *testing.T) {
	t.Parallel()
	fsys := New()
	err := fsys.WriteFile("myfile.txt", []byte("Hello, world!"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	data, err := fsys.ReadFile("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "Hello, world!" {
		t.Errorf("expected Hello, world!, got %s", string(data))
	}

	err = fsys.WriteFile("myfile.txt", []byte("New data"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	data, err = fsys.ReadFile("myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "New data" {
		t.Errorf("expected New data, got %s", string(data))
	}
}

func BenchmarkFSWriteFile(b *testing.B) {
	b.StopTimer()
	fsys := New()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err := fsys.WriteFile("myfile.txt", []byte(strings.Repeat("A", 1024)), 0644)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileStat(t *testing.T) {
	t.Parallel()
	now := time.Now()
	f := &File{name: "myfile.txt", mode: 0644, data: []byte("Hello"), modTime: now}
	fileInfo, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Name() != "myfile.txt" {
		t.Errorf("expected myfile.txt, got %s", fileInfo.Name())
	}
	if fileInfo.Size() != 5 {
		t.Errorf("expected 5, got %d", fileInfo.Size())
	}
	if fileInfo.Mode() != 0644 {
		t.Errorf("expected 0644, got %o", fileInfo.Mode())
	}
	if fileInfo.ModTime().IsZero() {
		t.Errorf("expected non-zero time, got zero time")
	}
	if fileInfo.IsDir() {
		t.Errorf("expected false, got true")
	}
	if fileInfo.Sys() != nil {
		t.Errorf("expected nil, got %v", fileInfo.Sys())
	}
}

func BenchmarkFileStat(b *testing.B) {
	f := &File{name: "myfile.txt", mode: 0644, data: []byte("Hello")}
	for i := 0; i < b.N; i++ {
		_, err := f.Stat()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileRead(t *testing.T) {
	t.Parallel()
	f := &File{data: []byte("Hello, world!")}
	b := make([]byte, 5)
	n, err := f.Read(b)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
	if string(b) != "Hello" {
		t.Errorf("expected Hello, got %s", string(b))
	}

	b = make([]byte, 10)
	n, err = f.Read(b)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if n != 8 {
		t.Errorf("expected 7 got %d", n)
	}
	if string(b[:n]) != ", world!" {
		t.Errorf("expected , world!, got %s", string(b[:n]))
	}

	b = make([]byte, 5)
	n, err = f.Read(b)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func BenchmarkFileRead(b *testing.B) {
	f := &File{data: []byte(strings.Repeat("A", 1024))}
	buf := make([]byte, 512)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.offset = 0
		for {
			_, err := f.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func TestFileClose(t *testing.T) {
	t.Parallel()
	f := &File{}
	err := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Check if offset is reset to 0
	if f.offset != 0 {
		t.Errorf("expected offset to be 0, got %d", f.offset)
	}
}

func BenchmarkFileClose(b *testing.B) {
	f := &File{}
	for i := 0; i < b.N; i++ {
		err := f.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFileName(t *testing.T) {
	t.Parallel()
	f := &File{name: "myfile.txt"}
	if f.Name() != "myfile.txt" {
		t.Errorf("expected myfile.txt, got %s", f.Name())
	}
}

func BenchmarkFileName(b *testing.B) {
	f := &File{name: "myfile.txt"}
	for i := 0; i < b.N; i++ {
		f.Name()
	}
}

func TestFileSize(t *testing.T) {
	t.Parallel()
	f := &File{data: []byte("Hello")}
	if f.Size() != 5 {
		t.Errorf("expected 5, got %d", f.Size())
	}
}

func BenchmarkFileSize(b *testing.B) {
	f := &File{data: []byte("Hello")}
	for i := 0; i < b.N; i++ {
		f.Size()
	}
}

func TestFileMode(t *testing.T) {
	t.Parallel()
	f := &File{mode: 0644}
	if f.Mode() != 0644 {
		t.Errorf("expected 0644, got %o", f.Mode())
	}
}

func BenchmarkFileMode(b *testing.B) {
	f := &File{mode: 0644}
	for i := 0; i < b.N; i++ {
		f.Mode()
	}
}

func TestFileModTime(t *testing.T) {
	t.Parallel()
	now := time.Now()
	f := &File{modTime: now}
	if f.ModTime() != now {
		t.Errorf("expected %v, got %v", now, f.ModTime())
	}
}

func BenchmarkFileModTime(b *testing.B) {
	now := time.Now()
	f := &File{modTime: now}
	for i := 0; i < b.N; i++ {
		f.ModTime()
	}
}

func TestFileIsDir(t *testing.T) {
	t.Parallel()
	f := &File{mode: fs.ModeDir}
	if !f.IsDir() {
		t.Errorf("expected true, got false")
	}

	f = &File{mode: 0644}
	if f.IsDir() {
		t.Errorf("expected false, got true")
	}
}

func BenchmarkFileIsDir(b *testing.B) {
	f := &File{mode: fs.ModeDir}
	for i := 0; i < b.N; i++ {
		f.IsDir()
	}
}

func TestFileSys(t *testing.T) {
	t.Parallel()
	f := &File{}
	if f.Sys() != nil {
		t.Errorf("expected nil, got %v", f.Sys())
	}
}

func BenchmarkFileSys(b *testing.B) {
	f := &File{}
	for i := 0; i < b.N; i++ {
		f.Sys()
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	if ErrExist == nil {
		t.Errorf("ErrExist is nil")
	}
	if ErrNotFound == nil {
		t.Errorf("ErrNotFound is nil")
	}
}

func TestDefaultFileModeAndDirMode(t *testing.T) {
	t.Parallel()
	if DefaultFileMode != 0644 {
		t.Errorf("DefaultFileMode is not 0644")
	}
	if DefaultDirMode != 0755 {
		t.Errorf("DefaultDirMode is not 0755")
	}
}

func TestMkdirAll(t *testing.T) {
	t.Parallel()
	fsys := New()
	err := os.MkdirAll("testdata/dir1/dir2", 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("testdata")

	err = fsys.MkdirAll("testdata/dir1/dir2", 0755)
	if err != nil {
		t.Fatal(err)
	}

	_, err = fsys.Stat("testdata/dir1/dir2")
	if err != nil {
		t.Fatal(err)
	}
}

func (self *FS) MkdirAll(path string, perm fs.FileMode) (err error) {
	parts := strings.Split(path, "/")
	currentPath := ""
	for _, part := range parts {
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}
		_, err := self.Stat(currentPath)
		if err != nil {
			err = self.Mkdir(currentPath, perm)
			if err != nil && err != ErrExist {
				return err
			}
		}
	}
	return nil
}

func BenchmarkMkdirAll(b *testing.B) {
	fsys := New()
	for i := 0; i < b.N; i++ {
		err := fsys.MkdirAll("testdata/dir1/dir2", 0755)
		if err != nil {
			b.Fatal(err)
		}
	}
}
