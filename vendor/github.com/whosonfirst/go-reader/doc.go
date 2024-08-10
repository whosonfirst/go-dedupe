// Example:
//
//	package main
//
//	import (
//	 	"context"
//		"github.com/whosonfirst/go-reader"
//		"io"
//	 	"os"
//	)
//
//	 func main() {
//	 	ctx := context.Background()
//	 	r, _ := reader.NewReader(ctx, "fs:///usr/local/data")
//	 	fh, _ := r.Read(ctx, "example.txt")
//	 	defer fh.Close()
//	 	io.Copy(os.Stdout, fh)
//	 }
//
// Package reader provides a common interface for reading from a variety of sources. It has the following interface:
//
//	type Reader interface {
//		Read(context.Context, string) (io.ReadSeekCloser, error)
//		ReaderURI(string) string
//	}
//
// Reader intstances are created either by calling a package-specific New{SOME_READER}Reader method or by invoking the
// reader.NewReader method passing in a context.Context instance and a URI specific to the reader class. For example:
//
//	r, _ := reader.NewReader(ctx, "fs:///usr/local/data")
//
// Custom reader packages implement the reader.Reader interface and register their availability by calling the reader.RegisterRegister
// method on initialization. For example:
//
//	func init() {
//
//		ctx := context.Background()
//
//		err = RegisterReader(ctx, "file", NewFileReader)
//
//	 	if err != nil {
//			panic(err)
//		}
//	}
package reader
