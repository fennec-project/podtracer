package podtracer

import (
	"io"
	"os"
)

type Writers struct {

	// Collection of writers to where send the stdout output
	StdoutWriters []io.Writer

	// Collection of writers to where send the stderr output
	StderrWriters []io.Writer

	// Closers stores the close methods for all open files
	// and connections to gracefully shutdown
	Closers []func() error

	// Writers options
	EnableStdout bool

	EnableStderr bool
}

// Stdout and Stderr are open Files pointing to the standard output and standard error file descriptors
// Any command line output and errors will be written
func (w *Writers) SetOSWriters() {

	// /dev/stdout
	if w.EnableStdout {
		w.StdoutWriters = append(w.StdoutWriters, os.Stdout)
	}
	// /dev/stderr
	if w.EnableStderr {
		w.StderrWriters = append(w.StderrWriters, os.Stderr)
	}
	return
}

// Setup file writers to store stdout and stderr
func (w *Writers) SetFileWriters(stdoutFilePath string, stderrFilePath string) error {

	// TODO: check if the file exists and give a warning saying
	// it will be overwritten. Implement continue/append/cancel
	// options.

	fileClosers := []func() error{}

	if stdoutFilePath != "" {
		stdoutFile, err := os.OpenFile(stdoutFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		w.StdoutWriters = append(w.StdoutWriters, stdoutFile)

		stdoutCloser := func() error {
			err := stdoutFile.Close()
			if err != nil {
				return err
			}
			return nil
		}

		fileClosers = append(fileClosers, stdoutCloser)
	}

	if stderrFilePath != "" {
		stderrFile, err := os.OpenFile(stderrFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		w.StderrWriters = append(w.StderrWriters, stderrFile)

		stderrCloser := func() error {
			err := stderrFile.Close()
			if err != nil {
				return err
			}
			return nil
		}

		fileClosers = append(fileClosers, stderrCloser)
	}
	w.Closers = fileClosers
	return nil
}

// CleanUp executes all closers for graceful shutdown
func (w *Writers) CleanUp() error {

	for _, f := range w.Closers {
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

//NEXT STEP DREAM:

// setup kafka writer
// setup elasticsearch writer
// setup cassandra writer
// setup wireshark writer
// setup jupyter-notebook writer
// neo4j ?
