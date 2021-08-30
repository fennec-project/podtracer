package podtracer

import (
	"io"
	"os"
)

type Writers struct {

	// Collection of writers to where send the stdout output.
	stdoutWriters []io.Writer

	// Collection of writers to where send the stderr output
	stderrWriters []io.Writer

	// Writers options
	EnableStdout bool

	EnableStderr bool
}

// Stdout and Stderr are open Files pointing to the standard output and standard error file descriptors.
// Any command line output and errors will be written.
func (w *Writers) setOSWriters() {

	// /dev/stdout
	w.stdoutWriters = append(w.stdoutWriters, os.Stdout)
	// /dev/stderr
	w.stderrWriters = append(w.stderrWriters, os.Stderr)

	return
}

// Setup file writers to store stdout and stderr
func (w *Writers) setFileWriters(stdoutFilePath string, stderrFilePath string) error {

	// TODO: check if the file exists and give a warning saying
	// it will be overwritten. Implement continue/append/cancel
	// options.

	if stdoutFilePath != "" {
		stdoutFile, err := os.OpenFile(stdoutFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		w.stdoutWriters = append(w.stdoutWriters, stdoutFile)
	}

	if stderrFilePath != "" {
		stderrFile, err := os.OpenFile(stderrFilePath, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		w.stderrWriters = append(w.stderrWriters, stderrFile)
	}

	return nil
}

//NEXT STEP DREAM:

// setup kafka writer
// setup elasticsearch writer
// setup cassandra writer
// setup wireshark writer
// setup jupyter-notebook writer
