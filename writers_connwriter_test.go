// Copyright (c) 2012 - Cloud Instruments Co. Ltd.
// 
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met: 
// 
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer. 
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution. 
// 
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package seelog

import (
	"testing"
	"time"
	"net"
	"syscall"
)

type testObj struct {
	Num int
	Str string
	Val float32
	Time time.Time
}

var (
	connWriterLog1 = "Testasasaafasf"
	connWriterLog2 = "fgrehgsnkmrgergerg[234%:dfsads:2]"
	connWriterLog3 = " 3242 3 24.df.we"
	connWriterLog = connWriterLog1 + connWriterLog2 + connWriterLog3
	obj = &testObj { 11, "sdfasd", 12.5, time.Now()  }
)

func TestConnWriter_ReconnectOnMessage(t *testing.T) {
	server, err := startTcpServer(t)
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Expect(connWriterLog)
	
	writer :=  newConnWriter("tcp4", ":" + server.port, true)
	defer writer.Close()
	
	_, err = writer.Write([]byte(connWriterLog))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Wait()
	server.Close()
}

func TestConnWriter_OneConnect(t *testing.T) {
	server, err := startTcpServer(t)
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Expect(connWriterLog)
	
	writer :=  newConnWriter("tcp4", ":" + server.port, false)
		
	_, err = writer.Write([]byte(connWriterLog1))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	_, err = writer.Write([]byte(connWriterLog2))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	_, err = writer.Write([]byte(connWriterLog3))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	writer.Close()
	
	server.Wait()
	server.Close()
}

func TestConnWriter_ReconnectOnMessage_WriteError(t *testing.T) {
	server, err := startTcpServer(t)
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Expect(connWriterLog)
	
	writer :=  newConnWriter("tcp4", ":" + server.port, true)
	defer writer.Close()
	
	_, err = writer.Write([]byte(connWriterLog))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Wait()
	server.Close()
	
	_, err = writer.Write([]byte(connWriterLog))
	if err == nil {
		t.Fatal("Write to closed server must return error")
		return
	}
	
	operr, ok := err.(*net.OpError)
	if !ok {
		t.Fatalf("Expected *net.OpError. Got: %v", err)
		return
	}
	
	errno, ok := operr.Err.(syscall.Errno)
	if !ok {
		t.Fatalf("Expected syscall.Errno. Got %v", operr)
		return
	}
	
	if errno != syscall.ECONNREFUSED {
		t.Fatalf("Expected syscall.ECONNREFUSED. Got %v", errno)
		return
	}
	
	err = server.Start()
		if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Expect(connWriterLog)
	
	_, err = writer.Write([]byte(connWriterLog))
	if err != nil {
		t.Fatal(err)
		return
	}
	
	server.Wait()
	server.Close()
}