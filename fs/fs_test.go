package fs

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var fsys = New()
	if fsys == nil {
		t.Errorf("New() returned nil")
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New()
	}
}

func TestFS_Open(t *testing.T) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var (
		f, openErr = fsys.Open(testFile)
	)
	if openErr != nil {
		t.Errorf("Open() error = %v, want nil", openErr)
		return
	}
	defer f.Close()

	file, ok := f.(*File)
	if !ok {
		t.Fatalf("Failed to cast fs.File to *File")
	}

	var (
		fileInfo, statErr = file.Stat()
	)
	if statErr != nil {
		t.Errorf("Stat() error = %v, want nil", statErr)
		return
	}

	if fileInfo.Name() != testFile {
		t.Errorf("Name() = %v, want %v", fileInfo.Name(), testFile)
	}

	var (
		readData  = make([]byte, len(testData))
		n, readErr = file.Read(readData)
	)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		t.Errorf("Read() error = %v, want nil or io.EOF", readErr)
		return
	}
	if n != len(testData) {
		t.Errorf("Read() read %v bytes, want %v", n, len(testData))
	}
	if string(readData) != string(testData) {
		t.Errorf("Read() data = %v, want %v", string(readData), string(testData))
	}

	_, err = fsys.Open("nonexistent.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Open() for nonexistent file, error = %v, want ErrNotFound", err)
	}
}

func BenchmarkFS_Open(b *testing.B) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			file, err = fsys.Open(testFile)
		)
		if err != nil {
			b.Errorf("Open() error = %v, want nil", err)
		}
		if file != nil {
			file.Close()
		}
	}
}

func TestFS_Stat(t *testing.T) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var (
		fileInfo, statErr = fsys.Stat(testFile)
	)
	if statErr != nil {
		t.Errorf("Stat() error = %v, want nil", statErr)
		return
	}

	if fileInfo.Name() != testFile {
		t.Errorf("Name() = %v, want %v", fileInfo.Name(), testFile)
	}
	if fileInfo.Size() != int64(len(testData)) {
		t.Errorf("Size() = %v, want %v", fileInfo.Size(), len(testData))
	}
	if fileInfo.Mode() != DefaultFileMode {
		t.Errorf("Mode() = %v, want %v", fileInfo.Mode(), DefaultFileMode)
	}
	if fileInfo.ModTime().IsZero() {
		t.Errorf("ModTime() is zero, want non-zero")
	}
	if fileInfo.IsDir() {
		t.Errorf("IsDir() = true, want false")
	}
	if fileInfo.Sys() != nil {
		t.Errorf("Sys() = %v, want nil", fileInfo.Sys())
	}

	_, err = fsys.Stat("nonexistent.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Stat() for nonexistent file, error = %v, want ErrNotFound", err)
	}
}

func BenchmarkFS_Stat(b *testing.B) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = fsys.Stat(testFile)
		if err != nil {
			b.Errorf("Stat() error = %v, want nil", err)
		}
	}
}

func TestFS_ReadFile(t *testing.T) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	var (
		readData, readErr = fsys.ReadFile(testFile)
	)
	if readErr != nil {
		t.Errorf("ReadFile() error = %v, want nil", readErr)
		return
	}

	if string(readData) != string(testData) {
		t.Errorf("ReadFile() data = %v, want %v", string(readData), string(testData))
	}

	_, err = fsys.ReadFile("nonexistent.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadFile() for nonexistent file, error = %v, want ErrNotFound", err)
	}
}

func BenchmarkFS_ReadFile(b *testing.B) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		_, err     = fsys.Create(testFile)
	)
	if err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}
	err = fsys.WriteFile(testFile, testData, DefaultFileMode)
	if err != nil {
		b.Fatalf("Failed to write test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = fsys.ReadFile(testFile)
		if err != nil {
			b.Errorf("ReadFile() error = %v, want nil", err)
		}
	}
}

func TestFS_ReadDir(t *testing.T) {
	var (
		fsys     = New()
		testDir    = "testdir"
		testFile1  = filepath.Join(testDir, "file1.txt")
		testFile2  = filepath.Join(testDir, "file2.txt")
		mkdirErr   = fsys.Mkdir(testDir, DefaultDirMode)
	)
	if mkdirErr != nil {
		t.Fatalf("Failed to create test directory: %v", mkdirErr)
	}
	var writeErr1 = fsys.WriteFile(testFile1, []byte("data1"), DefaultFileMode)
	if writeErr1 != nil {
		t.Fatalf("Failed to write test file 1: %v", writeErr1)
	}
	var writeErr2 = fsys.WriteFile(testFile2, []byte("data2"), DefaultFileMode)
	if writeErr2 != nil {
		t.Fatalf("Failed to write test file 2: %v", writeErr2)
	}

	var (
		entries, readDirErr = fsys.ReadDir(testDir)
	)
	if readDirErr != nil {
		t.Errorf("ReadDir() error = %v, want nil", readDirErr)
		return
	}

	if len(entries) != 2 {
		t.Errorf("ReadDir() len = %v, want %v", len(entries), 2)
	}

	var (
		names = []string{entries[0].Name(), entries[1].Name()}
	)
	if !strings.Contains(strings.Join(names, " "), "file1.txt") || !strings.Contains(strings.Join(names, " "), "file2.txt") {
		t.Errorf("ReadDir() entries = %v, want [file1.txt file2.txt]", names)
	}

	_, err := fsys.ReadDir("nonexistent")
	if err == nil {
		t.Errorf("ReadDir() for nonexistent dir, error = %v, want non-nil", err)
	}
}

func TestFS_ReadDir_NotFound(t *testing.T) {
	var fsys = New()
	_, err := fsys.ReadDir("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("ReadDir() for nonexistent dir, error = %v, want ErrNotFound", err)
	}
}

func BenchmarkFS_ReadDir(b *testing.B) {
	var (
		fsys     = New()
		testDir    = "testdir"
		testFile1  = filepath.Join(testDir, "file1.txt")
		testFile2  = filepath.Join(testDir, "file2.txt")
		mkdirErr = fsys.Mkdir(testDir, DefaultDirMode)
	)
	if mkdirErr != nil {
		b.Fatalf("Failed to create test directory: %v", mkdirErr)
	}
	var writeErr1 = fsys.WriteFile(testFile1, []byte("data1"), DefaultFileMode)
	if writeErr1 != nil {
		b.Fatalf("Failed to write test file 1: %v", writeErr1)
	}
	var writeErr2 = fsys.WriteFile(testFile2, []byte("data2"), DefaultFileMode)
	if writeErr2 != nil {
		b.Fatalf("Failed to write test file 2: %v", writeErr2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fsys.ReadDir(testDir)
		if err != nil {
			b.Errorf("ReadDir() error = %v, want nil", err)
		}
	}
}

func TestFS_Create(t *testing.T) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		file, err = fsys.Create(testFile)
	)
	if err != nil {
		t.Errorf("Create() error = %v, want nil", err)
		return
	}

	if file == nil {
		t.Errorf("Create() file = nil, want non-nil")
		return
	}

	if file.(*File).Name() != testFile {
		t.Errorf("Create() file.Name() = %v, want %v", file.(*File).Name(), testFile)
	}

	_, err = fsys.Create(testFile)
	if !errors.Is(err, ErrExist) {
		t.Errorf("Create() existing file, error = %v, want ErrExist", err)
	}
}

func BenchmarkFS_Create(b *testing.B) {
	var fsys = New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			testFile  = "testfile.txt"
			_, err = fsys.Create(testFile)
		)
		if err != nil && !errors.Is(err, ErrExist) {
			b.Errorf("Create() error = %v, want nil", err)
		}
		fsys.files.Delete(testFile) // Clean up for next iteration
	}
}

func TestFS_Mkdir(t *testing.T) {
	var (
		fsys    = New()
		testDir   = "testdir"
		mkdirErr = fsys.Mkdir(testDir, DefaultDirMode)
	)
	if mkdirErr != nil {
		t.Fatalf("Failed to create test directory: %v", mkdirErr)
	}

	var (
		fileInfo, statErr = fsys.Stat(testDir)
	)
	if statErr != nil {
		t.Errorf("Stat() error = %v, want nil", statErr)
		return
	}

	if !fileInfo.IsDir() {
		t.Errorf("IsDir() = false, want true")
	}

	err := fsys.Mkdir(testDir, DefaultDirMode)
	if !errors.Is(err, ErrExist) {
		t.Errorf("Mkdir() existing dir, error = %v, want ErrExist", err)
	}
}

func BenchmarkFS_Mkdir(b *testing.B) {
	var fsys = New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			testDir   = "testdir"
			err       = fsys.Mkdir(testDir, DefaultDirMode)
		)
		if err != nil && !errors.Is(err, ErrExist) {
			b.Errorf("Mkdir() error = %v, want nil", err)
		}
		fsys.files.Delete(testDir) // Clean up for next iteration
	}
}

func TestFS_WriteFile(t *testing.T) {
	var (
		fsys    = New()
		testFile  = "testfile.txt"
		testData  = []byte("test data")
		err       = fsys.WriteFile(testFile, testData, DefaultFileMode)
	)
	if err != nil {
		t.Errorf("WriteFile() error = %v, want nil", err)
		return
	}

	var (
		readData, readErr = fsys.ReadFile(testFile)
	)
	if readErr != nil {
		t.Errorf("ReadFile() error = %v, want nil", readErr)
		return
	}

	if string(readData) != string(testData) {
		t.Errorf("WriteFile() data = %v, want %v", string(readData), string(testData))
	}

	// Test overwriting the file
	var (
		newData = []byte("new data")
		err2    = fsys.WriteFile(testFile, newData, DefaultFileMode)
	)
	if err2 != nil {
		t.Errorf("WriteFile() overwrite error = %v, want nil", err2)
		return
	}

	var (
		readData2, readErr2 = fsys.ReadFile(testFile)
	)
	if readErr2 != nil {
		t.Errorf("ReadFile() error = %v, want nil", readErr2)
		return
	}

	if string(readData2) != string(newData) {
		t.Errorf("WriteFile() overwrite data = %v, want %v", string(readData2), string(newData))
	}
}

func BenchmarkFS_WriteFile(b *testing.B) {
	var fsys = New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			testFile  = "testfile.txt"
			testData  = []byte("test data")
			err       = fsys.WriteFile(testFile, testData, DefaultFileMode)
		)
		if err != nil {
			b.Errorf("WriteFile() error = %v, want nil", err)
		}
		fsys.files.Delete(testFile) // Clean up for next iteration
	}
}

func TestFile_Write(t *testing.T) {
	var (
		file    = &File{data: []byte{}}
		testData  = []byte("test data")
		n, err    = file.Write(testData)
	)
	if err != nil {
		t.Errorf("Write() error = %v, want nil", err)
		return
	}

	if n != len(testData) {
		t.Errorf("Write() n = %v, want %v", n, len(testData))
	}

	if string(file.data) != string(testData) {
		t.Errorf("Write() data = %v, want %v", string(file.data), string(testData))
	}
}

func BenchmarkFile_Write(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var (
			file    = &File{data: []byte{}}
			testData  = []byte("test data")
			_, err    = file.Write(testData)
		)
		if err != nil {
			b.Errorf("Write() error = %v, want nil", err)
		}
	}
}

func TestFile_WriteTo(t *testing.T) {
	var (
		file    = &File{data: []byte("test data")}
		buffer  strings.Builder
		n, err    = file.WriteTo(&buffer)
	)
	if err != nil {
		t.Errorf("WriteTo() error = %v, want nil", err)
		return
	}

	if n != int64(len(file.data)) {
		t.Errorf("WriteTo() n = %v, want %v", n, len(file.data))
	}

	if buffer.String() != string(file.data) {
		t.Errorf("WriteTo() buffer = %v, want %v", buffer.String(), string(file.data))
	}
}

func BenchmarkFile_WriteTo(b *testing.B) {
	var file = &File{data: []byte("test data")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buffer strings.Builder
		_, err := file.WriteTo(&buffer)
		if err != nil {
			b.Errorf("WriteTo() error = %v, want nil", err)
		}
	}
}

func TestFile_Seek(t *testing.T) {
	var (
		file    = &File{data: []byte("test data")}
		offset, err = file.Seek(5, io.SeekStart)
	)
	if err != nil {
		t.Errorf("Seek() error = %v, want nil", err)
		return
	}

	if offset != 5 {
		t.Errorf("Seek() offset = %v, want %v", offset, 5)
	}

	var (
		offset2, err2 = file.Seek(-2, io.SeekEnd)
	)
	if err2 != nil {
		t.Errorf("Seek() error = %v, want nil", err2)
		return
	}

	if offset2 != int64(len(file.data)-2) {
		t.Errorf("Seek() offset = %v, want %v", offset2, len(file.data)-2)
	}

	var (
		offset3, err3 = file.Seek(2, io.SeekCurrent)
	)
	if err3 != nil {
		t.Errorf("Seek() error = %v, want nil", err3)
		return
	}

	if offset3 != int64(len(file.data)) {
		t.Errorf("Seek() offset = %v, want %v", offset3, len(file.data))
	}

	_, err = file.Seek(-100, io.SeekStart)
	if err != nil && err.Error() != "negative position" {
		t.Errorf("Seek() negative position error = %v, want negative position", err)
	}

	_, err = file.Seek(0, 3)
	if err != nil && err.Error() != "invalid whence" {
		t.Errorf("Seek() invalid whence error = %v, want invalid whence", err)
	}
}

func BenchmarkFile_Seek(b *testing.B) {
	var file = &File{data: []byte("test data")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := file.Seek(5, io.SeekStart)
		if err != nil {
			b.Errorf("Seek() error = %v, want nil", err)
		}
	}
}

func TestFile_Stat(t *testing.T) {
	var file = &File{name: "testfile.txt", mode: DefaultFileMode, data: []byte("test data"), modTime: time.Now()}
	var (
		fileInfo, err = file.Stat()
	)
	if err != nil {
		t.Errorf("Stat() error = %v, want nil", err)
		return
	}

	if fileInfo.Name() != file.name {
		t.Errorf("Name() = %v, want %v", fileInfo.Name(), file.name)
	}
	if fileInfo.Size() != int64(len(file.data)) {
		t.Errorf("Size() = %v, want %v", fileInfo.Size(), len(file.data))
	}
	if fileInfo.Mode() != file.mode {
		t.Errorf("Mode() = %v, want %v", fileInfo.Mode(), file.mode)
	}
	if fileInfo.ModTime() != file.modTime {
		t.Errorf("ModTime() = %v, want %v", fileInfo.ModTime(), file.modTime)
	}
	if fileInfo.IsDir() {
		t.Errorf("IsDir() = true, want false")
	}
	if fileInfo.Sys() != nil {
		t.Errorf("Sys() = %v, want nil", fileInfo.Sys())
	}
}

func BenchmarkFile_Stat(b *testing.B) {
	var file = &File{name: "testfile.txt", mode: DefaultFileMode, data: []byte("test data"), modTime: time.Now()}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := file.Stat()
		if err != nil {
			b.Errorf("Stat() error = %v, want nil", err)
		}
	}
}

func TestFile_Read(t *testing.T) {
	var (
		file    = &File{data: []byte("test data")}
		buffer  = make([]byte, 5)
		n, err    = file.Read(buffer)
	)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("Read() error = %v, want nil or io.EOF", err)
		return
	}

	if n != len(buffer) {
		t.Errorf("Read() n = %v, want %v", n, len(buffer))
	}

	if string(buffer) != "test " {
		t.Errorf("Read() buffer = %v, want %v", string(buffer), "test ")
	}

	var (
		buffer2 = make([]byte, 10)
		n2, err2  = file.Read(buffer2)
	)
	if err2 != nil && !errors.Is(err2, io.EOF) {
		t.Errorf("Read() error = %v, want nil or io.EOF", err2)
		return
	}

	if n2 != 4 {
		t.Errorf("Read() n = %v, want %v", n2, 4)
	}

	if string(buffer2[:n2]) != "data" {
		t.Errorf("Read() buffer = %v, want %v", string(buffer2[:n2]), "data")
	}

	_, err = file.Read(buffer)
	if !errors.Is(err, io.EOF) {
		t.Errorf("Read() at EOF error = %v, want io.EOF", err)
	}
}

func BenchmarkFile_Read(b *testing.B) {
	var file = &File{data: []byte("test data")}
	var buffer = make([]byte, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file.offset = 0 // Reset offset for each iteration
		_, err := file.Read(buffer)
		if err != nil && !errors.Is(err, io.EOF) {
			b.Errorf("Read() error = %v, want nil or io.EOF", err)
		}
	}
}

func TestFile_Close(t *testing.T) {
	var file = &File{offset: 10}
	var err = file.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
		return
	}

	if file.offset != 0 {
		t.Errorf("Close() offset = %v, want 0", file.offset)
	}
}

func BenchmarkFile_Close(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var file = &File{offset: 10}
		err := file.Close()
		if err != nil {
			b.Errorf("Close() error = %v, want nil", err)
		}
	}
}

func TestFile_Name(t *testing.T) {
	var (
		file    = &File{name: "testfile.txt"}
		name    = file.Name()
	)
	if name != "testfile.txt" {
		t.Errorf("Name() = %v, want %v", name, "testfile.txt")
	}
}

func BenchmarkFile_Name(b *testing.B) {
	var file = &File{name: "testfile.txt"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.Name()
	}
}

func TestFile_Size(t *testing.T) {
	var (
		file    = &File{data: []byte("test data")}
		size    = file.Size()
	)
	if size != 9 {
		t.Errorf("Size() = %v, want %v", size, 9)
	}
}

func BenchmarkFile_Size(b *testing.B) {
	var file = &File{data: []byte("test data")}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.Size()
	}
}

func TestFile_Mode(t *testing.T) {
	var (
		file    = &File{mode: DefaultFileMode}
		mode    = file.Mode()
	)
	if mode != DefaultFileMode {
		t.Errorf("Mode() = %v, want %v", mode, DefaultFileMode)
	}
}

func BenchmarkFile_Mode(b *testing.B) {
	var file = &File{mode: DefaultFileMode}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.Mode()
	}
}

func TestFile_ModTime(t *testing.T) {
	var now = time.Now()
	var (
		file    = &File{modTime: now}
		modTime = file.ModTime()
	)
	if modTime != now {
		t.Errorf("ModTime() = %v, want %v", modTime, now)
	}
}

func BenchmarkFile_ModTime(b *testing.B) {
	var now = time.Now()
	var file = &File{modTime: now}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.ModTime()
	}
}

func TestFile_IsDir(t *testing.T) {
	var file = &File{mode: fs.ModeDir}
	var isDir = file.IsDir()
	if !isDir {
		t.Errorf("IsDir() = %v, want %v", isDir, true)
	}

	var file2 = &File{mode: DefaultFileMode}
	var isDir2 = file2.IsDir()
	if isDir2 {
		t.Errorf("IsDir() = %v, want %v", isDir2, false)
	}
}

func BenchmarkFile_IsDir(b *testing.B) {
	var file = &File{mode: fs.ModeDir}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.IsDir()
	}
}

func TestFile_Sys(t *testing.T) {
	var file = &File{}
	var sys = file.Sys()
	if sys != nil {
		t.Errorf("Sys() = %v, want %v", sys, nil)
	}
}

func BenchmarkFile_Sys(b *testing.B) {
	var file = &File{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = file.Sys()
	}
}

func TestDirEntry_Name(t *testing.T) {
	var de = &dirEntry{name: "test"}
	var name = de.Name()
	if name != "test" {
		t.Errorf("Name() = %v, want %v", name, "test")
	}
}

func BenchmarkDirEntry_Name(b *testing.B) {
	var de = &dirEntry{name: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = de.Name()
	}
}

func TestDirEntry_IsDir(t *testing.T) {
	var de = &dirEntry{isDir: true}
	var isDir = de.IsDir()
	if !isDir {
		t.Errorf("IsDir() = %v, want %v", isDir, true)
	}

	var de2 = &dirEntry{isDir: false}
	var isDir2 = de2.IsDir()
	if isDir2 {
		t.Errorf("IsDir() = %v, want %v", isDir2, false)
	}
}

func BenchmarkDirEntry_IsDir(b *testing.B) {
	var de = &dirEntry{isDir: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = de.IsDir()
	}
}

func TestDirEntry_Type(t *testing.T) {
	var de = &dirEntry{isDir: true}
	var typeVal = de.Type()
	if typeVal != fs.ModeDir {
		t.Errorf("Type() = %v, want %v", typeVal, fs.ModeDir)
	}

	var de2 = &dirEntry{isDir: false}
	var typeVal2 = de2.Type()
	if typeVal2 != 0 {
		t.Errorf("Type() = %v, want %v", typeVal2, 0)
	}
}

func BenchmarkDirEntry_Type(b *testing.B) {
	var de = &dirEntry{isDir: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = de.Type()
	}
}

func TestDirEntry_Info(t *testing.T) {
	var de = &dirEntry{name: "test"}
	var (
		info, err = de.Info()
	)
	if err != nil {
		t.Errorf("Info() error = %v, want nil", err)
		return
	}

	if info.Name() != de.name {
		t.Errorf("Info().Name() = %v, want %v", info.Name(), de.name)
	}
}

func BenchmarkDirEntry_Info(b *testing.B) {
	var de = &dirEntry{name: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := de.Info()
		if err != nil {
			b.Errorf("Info() error = %v, want nil", err)
		}
	}
}
