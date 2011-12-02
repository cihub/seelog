package dispatchers

import (
	"testing"
	. "github.com/cihub/sealog/common"
	. "github.com/cihub/sealog/test"
	"github.com/cihub/sealog/format"
)

func TestFormattedWriter(t *testing.T) {
	formatStr := "%Level %LEVEL %Msg"
	message := "message"
	var logLevel LogLevel =  TraceLvl
	
	bytesVerifier, err := NewBytesVerfier(t)
	if err != nil {
		t.Error(err)
		return
	}
	
	formatter, err := format.NewFormatter(formatStr)
	if err != nil {
		t.Error(err)
		return
	}
	
	writer, err := NewFormattedWriter(bytesVerifier, formatter)
	if err != nil {
		t.Error(err)
		return
	}
	
	context, err := CurrentContext()
	if err != nil {
		t.Error(err)
		return
	}
	
	logMessage := formatter.Format(message, logLevel, context)
	
	bytesVerifier.ExpectBytes([]byte(logMessage))
	writer.Write(message, logLevel, context)
	bytesVerifier.MustNotExpect()
}