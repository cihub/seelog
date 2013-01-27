// Copyright (c) 2012 - Cloud Instruments Co., Ltd.
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
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	portStart = 1200
	portEnd   = 1299
)

type tcpServer struct {
	testEnv      *testing.T
	expect       bool
	expectedText string

	serveConnLock *sync.Mutex

	//closeChan chan bool
	closed         bool
	readallTimeout time.Duration
	listener       *net.TCPListener
	port           string
}

func startTcpServer(testEnv *testing.T) (*tcpServer, error) {
	server := new(tcpServer)

	server.serveConnLock = &sync.Mutex{}

	//server.closeChan = make(chan bool, 0)
	server.readallTimeout = 1 * time.Second
	server.testEnv = testEnv

	return server, server.Start()
}

func (server *tcpServer) Start() error {
	port := portStart
	for {
		tcpAddr, err := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(port))
		if err != nil {
			return err
		}

		server.listener, err = net.ListenTCP("tcp", tcpAddr)
		if err == nil {
			break
		}

		if port > portEnd {
			return errors.New(fmt.Sprintf("Port number exceeds %v", portEnd))
		}

		port++
	}
	server.port = strconv.Itoa(port)

	go func() {
		for {
			acceptNext := server.acceptCycle()
			if !acceptNext {
				return
			}
		}
	}()

	return nil
}

func (server *tcpServer) acceptCycle() (acceptNext bool) {
	acceptNext = true

	conn, err := server.listener.Accept()
	if err != nil {
		if !server.closed {
			server.testEnv.Fatal(err)
		} else {
			acceptNext = false
		}
		return
	}

	server.serveConnLock.Lock()
	defer server.serveConnLock.Unlock()

	if !server.expect {
		server.testEnv.Fatal("Unexpected connection")
		conn.Close()
		return
	}
	server.expect = false

	inputText := ""
	c := make(chan string, 0)
	go server.readAll(c, conn)

	select {
	case inputText = <-c:
		{
		}
	case <-time.After(server.readallTimeout):
		server.testEnv.Fatal("Timeout on readAll")
		return
	}

	if server.expectedText != inputText {
		server.testEnv.Fatalf("Incorrect input. Expected: %v. Got: %v", server.expectedText, inputText)
		conn.Close()
		return
	}

	conn.Close()

	return
}

func (server *tcpServer) readAll(c chan string, conn net.Conn) {
	bytes, err := ioutil.ReadAll(conn)
	if err != nil {
		server.testEnv.Error(err)
		return
	}

	c <- string(bytes)
}

func (server *tcpServer) Expect(text string) {
	server.serveConnLock.Lock()
	defer server.serveConnLock.Unlock()

	server.expect = true
	server.expectedText = text
}

func (server *tcpServer) Wait() {
	// Waits server.listener.Accept()
	<-time.After(100 * time.Microsecond)

	server.serveConnLock.Lock()
	defer server.serveConnLock.Unlock()
}

func (server *tcpServer) Close() {
	server.serveConnLock.Lock()
	defer server.serveConnLock.Unlock()

	server.closed = true
	server.listener.Close()
}
