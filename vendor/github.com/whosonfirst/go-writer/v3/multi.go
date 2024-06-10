package writer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"

	chain "github.com/g8rswimmer/error-chain"
)

type MultiWriterOptions struct {
	Writers []Writer
	Async   bool
	Logger  *log.Logger
	Verbose bool
}

// Type MultiWriter implements the `Writer` interface for writing documents to multiple `Writer` instances.
type MultiWriter struct {
	Writer
	writers []Writer
	logger  *log.Logger
	async   bool
	verbose bool
}

// NewMultiWriter returns a Writer instance that will send all writes to each instance in 'writers'.
// Writes happen synchronolously in the order in which the underlying Writer instances are specified.
func NewMultiWriter(ctx context.Context, writers ...Writer) (Writer, error) {

	opts := &MultiWriterOptions{
		Writers: writers,
	}

	return NewMultiWriterWithOptions(ctx, opts)
}

// NewMultiWriter returns a Writer instance that will send all writes to each instance in 'writers' asynchronously.
func NewAsyncMultiWriter(ctx context.Context, writers ...Writer) (Writer, error) {

	opts := &MultiWriterOptions{
		Writers: writers,
		Async:   true,
	}

	return NewMultiWriterWithOptions(ctx, opts)
}

// NewMultiWriterWithOptions returns a Writer instance derived from 'opts'.
func NewMultiWriterWithOptions(ctx context.Context, opts *MultiWriterOptions) (Writer, error) {

	wr := &MultiWriter{
		writers: opts.Writers,
		async:   opts.Async,
		verbose: opts.Verbose,
	}

	var logger *log.Logger

	if opts.Logger != nil {
		logger = opts.Logger
	} else {
		logger = log.New(io.Discard, "", 0)
	}

	err := wr.SetLogger(ctx, logger)

	if err != nil {
		return nil, fmt.Errorf("Failed to set logger, %w", err)
	}

	return wr, nil
}

// Write copies the contents of 'fh' to each of the writers contained by 'mw' in the order they
// were specified unless 'mw' was created by `NewAsyncMultiWriter`.
func (mw *MultiWriter) Write(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	if mw.async {
		return mw.writeAsync(ctx, key, fh)
	}

	return mw.writeSync(ctx, key, fh)
}

func (mw *MultiWriter) writeSync(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	errors := make([]error, 0)
	count := int64(0)

	for _, wr := range mw.writers {

		i, err := wr.Write(ctx, key, fh)

		if err != nil {
			slog.Error("Failed to write (sync)", "writer", fmt.Sprintf("%T", wr), "key", key, "error", err)
			errors = append(errors, err)
			continue
		}

		count += i
		_, err = fh.Seek(0, 0)

		if err != nil {
			return count, err
		}

		if mw.verbose {
			slog.Debug("Write successful (async)", "writer", fmt.Sprintf("%T", wr), "key", key)
			mw.logger.Printf("Wrote %s to %T", key, wr)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Write operations failed")
		err = errorChain(err, errors...)

		return count, err
	}

	return count, nil
}

func (mw *MultiWriter) writeAsync(ctx context.Context, key string, fh io.ReadSeeker) (int64, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return 0, fmt.Errorf("Failed to read body, %w", err)
	}

	done_ch := make(chan bool)
	err_ch := make(chan error)
	count_ch := make(chan int64)

	for _, wr := range mw.writers {

		go func(ctx context.Context, wr Writer, key string, body []byte) {

			defer func() {
				done_ch <- true
			}()

			i, err := wr.Write(ctx, key, bytes.NewReader(body))

			if err != nil {
				slog.Error("Failed to write (async)", "writer", fmt.Sprintf("%T", wr), "key", key, "error", err)
				err_ch <- err
				return
			}

			count_ch <- i

			if mw.verbose {
				slog.Debug("Write successful (async)", "writer", fmt.Sprintf("%T", wr), "key", key)
				mw.logger.Printf("Wrote %s to %T", key, wr)
			}

		}(ctx, wr, key, body)
	}

	count := int64(0)
	errors := make([]error, 0)

	remaining := len(mw.writers)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			errors = append(errors, err)
		case i := <-count_ch:
			count += i
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Write operations failed")
		err = errorChain(err, errors...)

		return count, err
	}

	return count, nil
}

// WriteURI returns an empty string. Because 'mw' has multiple underlying `Writer` instances
// each of which specifies their own `WriteURI` methods it's either a choice of returning a
// concatenated string (with all the values) or an empty string. The decision was made to opt
// for the latter.
func (mw *MultiWriter) WriterURI(ctx context.Context, key string) string {
	return ""
}

// Flushes publishes any outstanding data for each of the underlying `Writer` instances (in the order they were specified
// to the 'mw' instance) unless 'mw' was created by `NewAsyncMultiWriter`.
func (mw *MultiWriter) Flush(ctx context.Context) error {

	if mw.async {
		return mw.flushAsync(ctx)
	}

	return mw.flushSync(ctx)
}

func (mw *MultiWriter) flushSync(ctx context.Context) error {

	errors := make([]error, 0)

	for _, wr := range mw.writers {

		err := wr.Flush(ctx)

		if err != nil {
			slog.Error("Failed to flush writer (sync)", "writer", fmt.Sprintf("%T", wr), "error", err)
			errors = append(errors, err)
		}

		if mw.verbose {
			slog.Debug("Flushed writer (sync)", "writer", fmt.Sprintf("%T", wr))
			mw.logger.Printf("Flushed on %T", wr)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Flush operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

func (mw *MultiWriter) flushAsync(ctx context.Context) error {

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, wr := range mw.writers {

		go func(ctx context.Context, wr Writer) {

			err := wr.Flush(ctx)

			if err != nil {
				slog.Error("Failed to flush writer (sync)", "writer", fmt.Sprintf("%T", wr), "error", err)
				err_ch <- err
			}

			done_ch <- true

			if mw.verbose {
				slog.Debug("Flushed writer (async)", "writer", fmt.Sprintf("%T", wr))
				mw.logger.Printf("Flushed on %T", wr)
			}

		}(ctx, wr)
	}

	errors := make([]error, 0)

	remaining := len(mw.writers)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Flush operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

// Closes closes each of the underlying `Writer` instances (in the order they were specified
// to the 'mw' instance) unless 'mw' was created by `NewAsyncMultiWriter`.
func (mw *MultiWriter) Close(ctx context.Context) error {

	if mw.async {
		return mw.closeAsync(ctx)
	}

	return mw.closeSync(ctx)
}

func (mw *MultiWriter) closeSync(ctx context.Context) error {

	errors := make([]error, 0)

	for _, wr := range mw.writers {

		err := wr.Close(ctx)

		if err != nil {
			slog.Error("Failed to close writer (sync)", "writer", fmt.Sprintf("%T", wr), "error", err)
			errors = append(errors, err)
		}

		if mw.verbose {
			slog.Debug("Closed writer successfully (sync)", "writer", fmt.Sprintf("%T", wr))
			mw.logger.Printf("Closed writer on %T", wr)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Close operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

func (mw *MultiWriter) closeAsync(ctx context.Context) error {

	errors := make([]error, 0)

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, wr := range mw.writers {

		go func(ctx context.Context, wr Writer) {

			err := wr.Close(ctx)

			if err != nil {
				slog.Error("Failed to close writer (sync)", "writer", fmt.Sprintf("%T", wr), "error", err)
				err_ch <- err
			}

			done_ch <- true

			if mw.verbose {
				slog.Debug("Closed writer successfully (async)", "writer", fmt.Sprintf("%T", wr))
				mw.logger.Printf("Closed writer on %T", wr)
			}

		}(ctx, wr)
	}

	remaining := len(mw.writers)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more Close operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

// SetLogger assign 'logger' to each of the underlying `Writer` instances (in the order they were specified
// to the 'mw' instance) unless 'mw' was created by `NewAsyncMultiWriter`.
func (mw *MultiWriter) SetLogger(ctx context.Context, logger *log.Logger) error {

	mw.logger = logger

	if mw.async {
		return mw.setLoggerAsync(ctx, logger)
	}

	return mw.setLoggerSync(ctx, logger)
}

func (mw *MultiWriter) setLoggerSync(ctx context.Context, logger *log.Logger) error {

	errors := make([]error, 0)

	for _, wr := range mw.writers {

		err := wr.SetLogger(ctx, logger)

		if err != nil {
			errors = append(errors, err)
		}

		if mw.verbose {
			mw.logger.Printf("Set logger on %T", wr)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more SetLogger operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

func (mw *MultiWriter) setLoggerAsync(ctx context.Context, logger *log.Logger) error {

	errors := make([]error, 0)

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, wr := range mw.writers {

		go func(ctx context.Context, wr Writer) {

			err := wr.SetLogger(ctx, logger)

			if err != nil {
				err_ch <- err
			}

			done_ch <- true

			if mw.verbose {
				mw.logger.Printf("Set logger on %T", wr)
			}

		}(ctx, wr)
	}

	remaining := len(mw.writers)

	for remaining > 0 {
		select {
		case <-done_ch:
			remaining -= 1
		case err := <-err_ch:
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {

		err := fmt.Errorf("One or more SetLogger operations failed")
		err = errorChain(err, errors...)

		return err
	}

	return nil
}

func errorChain(first error, others ...error) error {
	ec := chain.New()
	ec.Add(first)

	for _, e := range others {
		ec.Add(e)
	}

	return ec
}
