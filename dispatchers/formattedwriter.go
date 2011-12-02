// Package dispatcher implements log dispatching functionality.
// It allows to filter, duplicate, customize the flow of log streams.
package dispatchers

import (
	. "sealog/common"
	"sealog/format"
	"os"
	"io"
	"fmt"
)

type FormattedWriter struct {
	writer io.Writer
	formatter *format.Formatter
}

func NewFormattedWriter(writer io.Writer, formatter *format.Formatter) (*FormattedWriter, os.Error) {
	if formatter == nil {
		return nil, os.NewError("Formatter can not be nil")
	}
	
	return &FormattedWriter { writer, formatter }, nil
}

func (formattedWriter *FormattedWriter) Write(message string, level LogLevel, context *LogContext) os.Error {
	str := formattedWriter.formatter.Format(message, level, context)
	_, err := formattedWriter.writer.Write([]byte(str))
	return err
}

func (formattedWriter *FormattedWriter) String() string {
	return fmt.Sprintf("writer: %s, format: %s", formattedWriter.writer, formattedWriter.formatter)
}