// Copyright 2025 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package fs implements an in-memory file system.
//
// The FS struct represents a simple in-memory file system that
// implements the fs.FS interface. It provides methods for creating,
// opening, reading, and writing files and directories.
//
// Core Features:
//
//   - In-memory storage: Stores files and directories in memory.
//   - fs.FS interface implementation: Implements the standard Go file system interface.
//   - File and directory creation: Supports creating new files and directories.
//   - Read and write operations: Provides methods for reading and writing file contents.
//   - Directory listing: Allows listing the contents of directories.
//
// Usage:
//
//   - Use New to create a new FS instance.
//   - Use Create to create a new file.
//   - Use Mkdir to create a new directory.
//   - Use Open to open an existing file.
//   - Use ReadFile to read the contents of a file.
//   - Use WriteFile to write data to a file.
package fs

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/vedranvuk/ds/trie"
)

const (
	// DefaultFileMode is the default file mode assigned to files.
	DefaultFileMode = 0644
	// DefaultDirMode is the default file mode assigned to dir.
	DefaultDirMode = 0755
)

var (
	// ErrExist is returned when a file does not exist.
	ErrExist = errors.New("file exists")
	// ErrNotFound is returned when a file was not found.
	ErrNotFound = errors.New("not found")
)

// FS represents a simple in-memory file system.
//
// It uses a trie to store files and directories, allowing for efficient
// path-based lookups.
//
// Usage:
//
//   - Create a new FS instance using New.
//   - Add files and directories using Create and Mkdir.
//   - Access files and directories using Open, Stat, ReadFile, and ReadDir.
type FS struct {
	files *trie.Trie[*File]
}

// New creates a new, empty in-memory file system.
//
// Returns:
//
//   - out: A pointer to the newly created FS instance.
//
// Example:
//
//	fsys := New()
func New() (out *FS) {
	out = &FS{
		files: trie.New[*File](),
	}
	return
}

// Open returns a fs.File for the given name, or ErrNotFound if the file does not exist.
//
// Parameters:
//
//   - name: The name of the file to open.
//
// Returns:
//
//   - file: The opened file, or nil if an error occurred.
//   - err: An error, if any. Returns ErrNotFound if the file does not exist.
//
// Example:
//
//	fsys := New()
//	_, err := fsys.Create("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	file, err := fsys.Open("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	defer file.Close()
func (self *FS) Open(name string) (file fs.File, err error) {
	var (
		f      *File
		exists bool
	)
	if f, exists = self.files.Get(name); !exists {
		return nil, ErrNotFound
	}
	return f, nil
}

// Stat returns a fs.FileInfo describing the named file, or an error if the file does not exist.
//
// Parameters:
//
//   - name: The name of the file to stat.
//
// Returns:
//
//   - fileInfo: The FileInfo of the file, or nil if an error occurred.
//   - err: An error, if any. Returns ErrNotFound if the file does not exist.
//
// Example:
//
//	fsys := New()
//	_, err := fsys.Create("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	fileInfo, err := fsys.Stat("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(fileInfo.Name())
//	// Output: myfile.txt
func (self *FS) Stat(name string) (fileInfo fs.FileInfo, err error) {
	var file fs.File
	if file, err = self.Open(name); err != nil {
		return nil, err
	}
	return file.Stat()
}

// ReadFile reads the named file and returns the contents.
// It returns an error if the file does not exist or if another error occurs during the read.
//
// Parameters:
//
//   - name: The name of the file to read.
//
// Returns:
//
//   - data: The contents of the file.
//   - err: An error, if any. Returns ErrNotFound if the file does not exist.
//
// Example:
//
//	fsys := New()
//	err := fsys.WriteFile("myfile.txt", []byte("Hello, world!"), 0644)
//	if err != nil {
//		panic(err)
//	}
//	data, err := fsys.ReadFile("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(string(data))
//	// Output: Hello, world!
func (self *FS) ReadFile(name string) (data []byte, err error) {
	var f fs.File
	if f, err = self.Open(name); err != nil {
		return nil, err
	}

	var file, ok = f.(*File)
	if !ok {
		return nil, errors.New("bug: entry is not a *File")
	}
	defer file.Close()

	data = make([]byte, len(file.data))
	var n int
	if n, err = file.Read(data); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
		err = nil
	}
	data = data[:n]

	return
}

// ReadDir reads the directory named by dirname and returns
// a list of directory entries.
//
// Parameters:
//
//   - name: The name of the directory to read.
//
// Returns:
//
//   - entries: A list of directory entries.
//   - err: An error, if any.
//
// Example:
//
//	fsys := New()
//	fsys.Mkdir("mydir", 0755)
//	fsys.WriteFile("mydir/myfile.txt", []byte(""), 0644)
//	entries, err := fsys.ReadDir("mydir")
//	if err != nil {
//		panic(err)
//	}
//	for _, entry := range entries {
//		fmt.Println(entry.Name())
//	}
//	// Output: myfile.txt
func (self *FS) ReadDir(name string) (entries []fs.DirEntry, err error) {
	var result []fs.DirEntry

	self.files.Enum(func(key string, value *File) bool {
		dir := filepath.Dir(key)

		if name == "." {
			if !strings.Contains(key, "/") {
				result = append(result, &dirEntry{name: key, isDir: value.mode.IsDir()})
			}
		} else if dir == name {
			base := filepath.Base(key)
			result = append(result, &dirEntry{name: base, isDir: value.mode.IsDir()})
		} else if strings.HasPrefix(key, name+"/") {
			rel, err := filepath.Rel(name, key)
			if err != nil {
				return true
			}
			parts := strings.Split(rel, "/")
			if len(parts) > 0 {
				firstPart := parts[0]
				alreadyAdded := false
				for _, entry := range result {
					if entry.Name() == firstPart {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					var isDir bool
					if f, ok := self.files.Get(filepath.Join(name, firstPart)); ok {
						isDir = f.mode.IsDir()
					}
					result = append(result, &dirEntry{name: firstPart, isDir: isDir})
				}
			}
		}
		return true
	})

	return result, nil
}

// dirEntry holds info about a directory entry returned by [FS.ReadDir].
type dirEntry struct {
	name  string
	isDir bool
}

// Name returns the name of the file or directory.
//
// Example:
//
//	de := dirEntry{name: "myfile.txt"}
//	fmt.Println(de.Name())
//	// Output: myfile.txt
func (self *dirEntry) Name() string { return self.name }

// IsDir reports whether the entry describes a directory.
//
// Example:
//
//	de := dirEntry{isDir: true}
//	fmt.Println(de.IsDir())
//	// Output: true
func (self *dirEntry) IsDir() bool { return self.isDir }

// Type returns the type bits for the entry.
// The type bits are a subset of the standard FileMode bits.
//
// Example:
//
//	de := dirEntry{isDir: true}
//	fmt.Println(de.Type() == fs.ModeDir)
//	// Output: true
func (self *dirEntry) Type() fs.FileMode {
	if self.IsDir() {
		return fs.ModeDir
	}
	return 0
}

// Info returns the FileInfo for the file or directory described by the entry.
//
// Example:
//
//	de := dirEntry{name: "myfile.txt"}
//	info, err := de.Info()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(info.Name())
//	// Output:
func (self *dirEntry) Info() (fileInfo fs.FileInfo, err error) {
	// This would require a lookup in the FS to get the actual FileInfo.
	// For simplicity, we return a basic FileInfo here.
	return &File{name: self.name, mode: 0, data: []byte{}, offset: 0, modTime: time.Now()}, nil
}

// Create creates a new file in the FS. If the file already exists, it returns ErrExist.
//
// Parameters:
//
//   - name: The name of the file to create.
//
// Returns:
//
//   - file: The newly created file.
//   - err: An error, if any. Returns ErrExist if the file already exists.
//
// Example:
//
//	fsys := New()
//	file, err := fsys.Create("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(file.Name())
//	// Output: myfile.txt
func (self *FS) Create(name string) (file fs.File, err error) {
	_, exists := self.files.Get(name)
	if exists {
		return nil, ErrExist
	}

	newFile := &File{
		name:    name,
		mode:    DefaultFileMode,
		data:    []byte{},
		modTime: time.Now(),
	}

	self.files.Put(name, newFile)
	return newFile, nil
}

// Mkdir creates a new directory with the specified name and permissions.
// If a file or directory with the same name already exists, it returns ErrExist.
//
// Parameters:
//
//   - name: The name of the directory to create.
//   - perm: The permissions to use for the new directory.
//
// Returns:
//
//   - err: An error, if any. Returns ErrExist if the directory already exists.
//
// Example:
//
//	fsys := New()
//	err := fsys.Mkdir("mydir", 0755)
//	if err != nil {
//		panic(err)
//	}
//	fileInfo, err := fsys.Stat("mydir")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(fileInfo.IsDir())
//	// Output: true
func (self *FS) Mkdir(name string, perm fs.FileMode) (err error) {
	_, exists := self.files.Get(name)
	if exists {
		return ErrExist
	}

	newFile := &File{
		name:    name,
		mode:    fs.ModeDir | perm,
		data:    []byte{},
		modTime: time.Now(),
	}

	self.files.Put(name, newFile)
	return nil
}

// WriteFile writes data to a file, creating it if it doesn't exist, with specified permissions.
//
// Parameters:
//
//   - name: The name of the file to write to.
//   - data: The data to write to the file.
//   - perm: The file mode to set on the file.
//
// Returns:
//
//   - err: An error, if any.
//
// Example:
//
//	fsys := New()
//	err := fsys.WriteFile("myfile.txt", []byte("Hello, world!"), 0644)
//	if err != nil {
//		panic(err)
//	}
//	data, err := fsys.ReadFile("myfile.txt")
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(string(data))
//	// Output: Hello, world!
func (self *FS) WriteFile(name string, data []byte, perm fs.FileMode) (err error) {
	var f, exists = self.files.Get(name)
	if !exists {
		var newFile fs.File
		newFile, err = self.Create(name)
		if err != nil {
			return err
		}

		if fi, ok := newFile.(*File); !ok {
			return errors.New("type assertion to *File failed")
		} else {
			f = fi
		}
	}

	f.data = data
	f.mode = perm
	f.modTime = time.Now()

	return nil
}

// File represents a file in the in-memory file system.
type File struct {
	name    string
	mode    fs.FileMode
	data    []byte
	offset  int
	modTime time.Time
}

// Stat returns a FileInfo describing the file.
//
// Returns:
//
//   - fileInfo: The FileInfo of the file.
//   - err: An error, if any.
//
// Example:
//
//	f := &File{name: "myfile.txt"}
//	fileInfo, err := f.Stat()
//	if err != nil {
//		panic(err)
//	}
//	fmt.Println(fileInfo.Name())
//	// Output: myfile.txt
func (self *File) Stat() (fileInfo fs.FileInfo, err error) { return self, nil }

// Read reads up to len(b) bytes from the File.
// It returns the number of bytes read and an error, if any.
// When the end of the file is reached, Read returns 0, io.EOF.
//
// Parameters:
//
//   - b: The byte slice to read into.
//
// Returns:
//
//   - n: The number of bytes read.
//   - err: An error, if any. Returns io.EOF if the end of the file is reached.
//
// Example:
//
//	f := &File{data: []byte("Hello, world!")}
//	b := make([]byte, 5)
//	n, err := f.Read(b)
//	if err != nil && err.Error() != "EOF" {
//		panic(err)
//	}
//	fmt.Println(string(b[:n]))
//	// Output: Hello
func (self *File) Read(b []byte) (n int, err error) {
	if self.offset >= len(self.data) {
		return 0, io.EOF
	}

	n = copy(b, self.data[self.offset:])
	self.offset += n

	if self.offset >= len(self.data) {
		return n, io.EOF
	}

	return n, nil
}

// Close closes the File, so it cannot be used for I/O anymore.
//
// Returns:
//
//   - err: An error, if any.
//
// Example:
//
//	f := &File{}
//	err := f.Close()
//	if err != nil {
//		panic(err)
//	}
func (self *File) Close() (err error) {
	self.offset = 0
	return nil
}

// Name returns the name of the file.
//
// Returns:
//
//   - name: The name of the file.
//
// Example:
//
//	f := &File{name: "myfile.txt"}
//	fmt.Println(f.Name())
//	// Output: myfile.txt
func (self *File) Name() string { return self.name }

// Size returns the length in bytes for regular files; other files may return different values.
//
// Returns:
//
//   - size: The length in bytes for regular files.
//
// Example:
//
//	f := &File{data: []byte("Hello")}
//	fmt.Println(f.Size())
//	// Output: 5
func (self *File) Size() int64 { return int64(len(self.data)) }

// Mode returns file mode bits.
//
// Returns:
//
//   - mode: The file mode bits.
//
// Example:
//
//	f := &File{mode: 0644}
//	fmt.Printf("%o\n", f.Mode())
//	// Output: 644
func (self *File) Mode() fs.FileMode { return self.mode }

// ModTime returns the modification time.
//
// Returns:
//
//   - modTime: The modification time.
//
// Example:
//
//	now := time.Now()
//	f := &File{modTime: now}
//	fmt.Println(f.ModTime().Format(time.RFC3339))
//	// Output: 2024-10-27T10:00:00+00:00 (example output, time will vary)
func (self *File) ModTime() time.Time { return self.modTime }

// IsDir reports whether the file is a directory.
//
// Returns:
//
//   - isDir: True if the file is a directory, false otherwise.
//
// Example:
//
//	f := &File{mode: fs.ModeDir}
//	fmt.Println(f.IsDir())
//	// Output: true
func (self *File) IsDir() bool { return self.mode.IsDir() }

// Sys is designed to return the underlying data source.
// Always returns nil in this implementation.
//
// Returns:
//
//   - nil.
func (self *File) Sys() interface{} { return nil }
