// Copyright 2011 Cloud Instruments Co. Ltd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dispatchers

import (
	"errors"
	"fmt"
	. "github.com/cihub/seelog/common"

	"github.com/cihub/seelog/format"
	"io"
)

type FormattedWriter struct {
	writer    io.Writer
	formatter *format.Formatter
}

func NewFormattedWriter(writer io.Writer, formatter *format.Formatter) (*FormattedWriter, error) {
	if formatter == nil {
		return nil, errors.New("Formatter can not be nil")
	}

	return &FormattedWriter{writer, formatter}, nil
}

func (formattedWriter *FormattedWriter) Write(message string, level LogLevel, context *LogContext) error {
	str := formattedWriter.formatter.Format(message, level, context)
	_, err := formattedWriter.writer.Write([]byte(str))
	return err
}

func (formattedWriter *FormattedWriter) String() string {
	return fmt.Sprintf("writer: %s, format: %s", formattedWriter.writer, formattedWriter.formatter)
}

func (formattedWriter *FormattedWriter) Writer() io.Writer {
	return formattedWriter.writer
}

func (formattedWriter *FormattedWriter) Format() *format.Formatter {
	return formattedWriter.formatter
}
