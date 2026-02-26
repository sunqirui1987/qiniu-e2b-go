package e2b

import (
	"context"
	"fmt"
	"time"
)

// FileType represents the type of a filesystem object
type FileType string

const (
	// FileTypeFile is a regular file
	FileTypeFile FileType = "file"
	// FileTypeDir is a directory
	FileTypeDir FileType = "dir"
)

// EntryInfo represents information about a filesystem object
// Similar to JS SDK's EntryInfo
type EntryInfo struct {
	// Name of the filesystem object
	Name string `json:"name"`
	// Type of the filesystem object (file or dir)
	Type FileType `json:"type"`
	// Path to the filesystem object
	Path string `json:"path"`
	// Size of the filesystem object in bytes
	Size int64 `json:"size"`
	// File mode and permission bits
	Mode int `json:"mode"`
	// String representation of file permissions (e.g. 'rwxr-xr-x')
	Permissions string `json:"permissions"`
	// Owner of the filesystem object
	Owner string `json:"owner"`
	// Group owner of the filesystem object
	Group string `json:"group"`
	// Last modification time
	ModifiedTime *time.Time `json:"modifiedTime,omitempty"`
}

// WriteInfo represents information about a written file
// Similar to JS SDK's WriteInfo
type WriteInfo struct {
	// Name of the filesystem object
	Name string `json:"name"`
	// Type of the filesystem object
	Type FileType `json:"type,omitempty"`
	// Path to the filesystem object
	Path string `json:"path"`
}

// Filesystem represents a module for interacting with the sandbox filesystem
// Similar to JS SDK's Filesystem class
type Filesystem struct {
	client    *Client
	sandboxID string
	localMode bool
	
	// Local filesystem storage for local mode
	localFiles map[string]*localFile
}

// localFile represents a file in local mode
type localFile struct {
	data     []byte
	isDir    bool
	mode     int
	modTime  time.Time
}

// NewFilesystem creates a new Filesystem instance
func NewFilesystem(client *Client, sandboxID string, localMode bool) *Filesystem {
	fs := &Filesystem{
		client:     client,
		sandboxID:  sandboxID,
		localMode:  localMode,
		localFiles: make(map[string]*localFile),
	}
	
	// Initialize root directory for local mode
	if localMode {
		fs.localFiles["/"] = &localFile{
			isDir:   true,
			mode:    0755,
			modTime: time.Now(),
		}
	}
	
	return fs
}

// Read reads file content as a string
// Similar to JS SDK's files.read(path)
func (fs *Filesystem) Read(path string) (string, error) {
	return fs.ReadWithContext(context.Background(), path)
}

// ReadWithContext reads file content as a string with context
func (fs *Filesystem) ReadWithContext(ctx context.Context, path string) (string, error) {
	if fs.localMode {
		return fs.readLocal(path)
	}
	
	data, err := fs.client.ReadFile(ctx, fs.sandboxID, path)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}

// ReadBytes reads file content as bytes
func (fs *Filesystem) ReadBytes(path string) ([]byte, error) {
	return fs.ReadBytesWithContext(context.Background(), path)
}

// ReadBytesWithContext reads file content as bytes with context
func (fs *Filesystem) ReadBytesWithContext(ctx context.Context, path string) ([]byte, error) {
	if fs.localMode {
		file, exists := fs.localFiles[path]
		if !exists {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		if file.isDir {
			return nil, fmt.Errorf("cannot read directory as file: %s", path)
		}
		return file.data, nil
	}
	
	return fs.client.ReadFile(ctx, fs.sandboxID, path)
}

// Write writes content to a file
// Similar to JS SDK's files.write(path, data)
func (fs *Filesystem) Write(path string, data []byte) error {
	return fs.WriteWithContext(context.Background(), path, data)
}

// WriteWithContext writes content to a file with context
func (fs *Filesystem) WriteWithContext(ctx context.Context, path string, data []byte) error {
	if fs.localMode {
		return fs.writeLocal(path, data, false)
	}
	
	return fs.client.WriteFile(ctx, fs.sandboxID, &WriteFileRequest{
		Path: path,
		Data: data,
	})
}

// WriteString writes string content to a file
func (fs *Filesystem) WriteString(path string, content string) error {
	return fs.Write(path, []byte(content))
}

// List lists entries in a directory
// Similar to JS SDK's files.list(path)
func (fs *Filesystem) List(path string) ([]*EntryInfo, error) {
	return fs.ListWithContext(context.Background(), path)
}

// ListWithContext lists entries in a directory with context
func (fs *Filesystem) ListWithContext(ctx context.Context, path string) ([]*EntryInfo, error) {
	if fs.localMode {
		return fs.listLocal(path)
	}
	
	files, err := fs.client.ListFiles(ctx, fs.sandboxID, path)
	if err != nil {
		return nil, err
	}
	
	// Convert File to EntryInfo
	entries := make([]*EntryInfo, len(files))
	for i, f := range files {
		fileType := FileTypeFile
		if f.IsDir {
			fileType = FileTypeDir
		}
		entries[i] = &EntryInfo{
			Name: f.Name,
			Type: fileType,
			Path: f.Path,
			Size: f.Size,
			Mode: f.Mode,
		}
	}
	
	return entries, nil
}

// MakeDir creates a new directory
// Similar to JS SDK's files.makeDir(path)
func (fs *Filesystem) MakeDir(path string) error {
	return fs.MakeDirWithContext(context.Background(), path)
}

// MakeDirWithContext creates a new directory with context
func (fs *Filesystem) MakeDirWithContext(ctx context.Context, path string) error {
	if fs.localMode {
		return fs.makeDirLocal(path)
	}
	
	return fs.client.MakeDir(ctx, fs.sandboxID, &MakeDirRequest{
		Path:      path,
		Recursive: true,
	})
}

// Remove removes a file or directory
// Similar to JS SDK's files.remove(path)
func (fs *Filesystem) Remove(path string) error {
	return fs.RemoveWithContext(context.Background(), path)
}

// RemoveWithContext removes a file or directory with context
func (fs *Filesystem) RemoveWithContext(ctx context.Context, path string) error {
	if fs.localMode {
		delete(fs.localFiles, path)
		return nil
	}
	
	return fs.client.RemoveFile(ctx, fs.sandboxID, path)
}

// Exists checks if a file or directory exists
// Similar to JS SDK's files.exists(path)
func (fs *Filesystem) Exists(path string) (bool, error) {
	return fs.ExistsWithContext(context.Background(), path)
}

// ExistsWithContext checks if a file or directory exists with context
func (fs *Filesystem) ExistsWithContext(ctx context.Context, path string) (bool, error) {
	if fs.localMode {
		_, exists := fs.localFiles[path]
		return exists, nil
	}
	
	// Try to get info
	_, err := fs.GetInfoWithContext(ctx, path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetInfo gets information about a file or directory
// Similar to JS SDK's files.getInfo(path)
func (fs *Filesystem) GetInfo(path string) (*EntryInfo, error) {
	return fs.GetInfoWithContext(context.Background(), path)
}

// GetInfoWithContext gets information about a file or directory with context
func (fs *Filesystem) GetInfoWithContext(ctx context.Context, path string) (*EntryInfo, error) {
	if fs.localMode {
		return fs.getInfoLocal(path)
	}
	
	// List parent directory and find the entry
	parent := getParentPath(path)
	entries, err := fs.ListWithContext(ctx, parent)
	if err != nil {
		return nil, err
	}
	
	name := getFileName(path)
	for _, entry := range entries {
		if entry.Name == name {
			return entry, nil
		}
	}
	
	return nil, fmt.Errorf("file not found: %s", path)
}

// Local mode implementations

func (fs *Filesystem) readLocal(path string) (string, error) {
	file, exists := fs.localFiles[path]
	if !exists {
		return "", fmt.Errorf("file not found: %s", path)
	}
	if file.isDir {
		return "", fmt.Errorf("cannot read directory as file: %s", path)
	}
	return string(file.data), nil
}

func (fs *Filesystem) writeLocal(path string, data []byte, isDir bool) error {
	fs.localFiles[path] = &localFile{
		data:    data,
		isDir:   isDir,
		mode:    0644,
		modTime: time.Now(),
	}
	return nil
}

func (fs *Filesystem) listLocal(path string) ([]*EntryInfo, error) {
	entries := []*EntryInfo{}
	
	for p, file := range fs.localFiles {
		if getParentPath(p) == path {
			fileType := FileTypeFile
			if file.isDir {
				fileType = FileTypeDir
			}
			entries = append(entries, &EntryInfo{
				Name: getFileName(p),
				Type: fileType,
				Path: p,
				Size: int64(len(file.data)),
				Mode: file.mode,
			})
		}
	}
	
	return entries, nil
}

func (fs *Filesystem) makeDirLocal(path string) error {
	fs.localFiles[path] = &localFile{
		isDir:   true,
		mode:    0755,
		modTime: time.Now(),
	}
	return nil
}

func (fs *Filesystem) getInfoLocal(path string) (*EntryInfo, error) {
	file, exists := fs.localFiles[path]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	
	fileType := FileTypeFile
	if file.isDir {
		fileType = FileTypeDir
	}
	
	return &EntryInfo{
		Name: getFileName(path),
		Type: fileType,
		Path: path,
		Size: int64(len(file.data)),
		Mode: file.mode,
	}, nil
}

// Helper functions

func getParentPath(path string) string {
	if path == "/" {
		return "/"
	}
	
	// Remove trailing slash
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	
	// Find last slash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	
	return "/"
}

func getFileName(path string) string {
	// Remove trailing slash
	for len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	
	// Find last slash
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	
	return path
}
