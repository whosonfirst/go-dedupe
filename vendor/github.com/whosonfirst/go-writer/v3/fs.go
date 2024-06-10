package writer

import (
	"context"
	"fmt"
	"github.com/natefinch/atomic"
	"io"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
)

// FileWriter is a struct that implements the `Writer` interface for writing documents as files on a local disk.
type FileWriter struct {
	Writer
	root      string
	dir_mode  os.FileMode
	file_mode os.FileMode
}

func init() {

	ctx := context.Background()

	schemes := []string{
		"fs",
	}

	for _, scheme := range schemes {

		err := RegisterWriter(ctx, scheme, NewFileWriter)

		if err != nil {
			panic(err)
		}
	}
}

// NewFileWriter returns a new `FileWriter` instance for writing documents as files on a local disk,
// configured by 'uri' in the form of:
//
//	fs://{PATH}
//
// Where {PATH} is an absolute path to an existing directory where files will be written.
func NewFileWriter(ctx context.Context, uri string) (Writer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse uri, %w", err)
	}

	root := u.Path
	info, err := os.Stat(root)

	if err != nil {
		return nil, fmt.Errorf("Failed to stat '%s', %w", root, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("root (%s) is not a directory", root)
	}

	// check for dir/file mode query parameters here

	wr := &FileWriter{
		dir_mode:  0755,
		file_mode: 0644,
		root:      root,
	}

	return wr, nil
}

// Write copies the content of 'fh' to 'path', where 'path' is assumed to be relative to the root
// path defined in the constuctor. If the root directory for 'path' does not exist it will be created.
func (wr *FileWriter) Write(ctx context.Context, path string, fh io.ReadSeeker) (int64, error) {

	abs_path := wr.WriterURI(ctx, path)
	abs_root := filepath.Dir(abs_path)

	_, err := os.Stat(abs_root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(abs_root, wr.dir_mode)

		if err != nil {
			return 0, fmt.Errorf("Failed to create %s, %w", abs_root, err)
		}
	}

	// So this... we can't do this because of cross-device (filesystem) limitations
	// under Unix. That is /tmp may be on a different filesystem than abs_path
	// tmp_file, err := os.CreateTemp("", filepath.Base(abs_path))
	// tmp_path := tmp_file.Name()

	tmp_suffix := fmt.Sprintf("tmp%d", rand.Int63())
	tmp_path := fmt.Sprintf("%s.%s", abs_path, tmp_suffix)

	tmp_file, err := os.OpenFile(tmp_path, os.O_RDWR|os.O_CREATE, 0600)

	if err != nil {
		return 0, fmt.Errorf("Failed to open temp file (%s), %w", tmp_path, err)
	}

	defer os.Remove(tmp_path)

	b, err := io.Copy(tmp_file, fh)

	if err != nil {
		return 0, fmt.Errorf("Failed to copy data to temp file (%s), %w", tmp_path, err)
	}

	err = tmp_file.Close()

	if err != nil {
		return 0, fmt.Errorf("Failed to close temp file (%s), %w", tmp_path, err)
	}

	err = os.Chmod(tmp_path, wr.file_mode)

	if err != nil {
		return 0, fmt.Errorf("Failed to assign permissions for temp file (%s), %w", tmp_path, err)
	}

	err = atomic.ReplaceFile(tmp_path, abs_path)

	if err != nil {
		return 0, fmt.Errorf("Failed to install final file (%s) from temp file (%s), %w", abs_path, tmp_path, err)
	}

	return b, nil
}

// WriterURI returns the absolute URL for 'path' relative to the root directory defined
// in the `FileWriter` constuctor.
func (wr *FileWriter) WriterURI(ctx context.Context, path string) string {
	return filepath.Join(wr.root, path)
}

// Flush is a no-op and returns nil.
func (wr *FileWriter) Flush(ctx context.Context) error {
	return nil
}

// Close closes the underlying writer mechanism.
func (wr *FileWriter) Close(ctx context.Context) error {
	return nil
}

// SetLogger assigns 'logger' to 'wr'.
func (wr *FileWriter) SetLogger(ctx context.Context, logger *log.Logger) error {
	return nil
}
