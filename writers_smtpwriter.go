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
	"net/smtp"
	"fmt"
)

const (
	subjectPhrase = "Diagnostic message from server: "
)

// smtpWriter is used to send emails via given SMTP-server.
type smtpWriter struct {
	auth smtp.Auth
	hostNameWithPort string
	senderAddress string
	senderName string
	recipientAddresses []string
}

// newSmtpWriter returns a new SMTP-writer
func newSmtpWriter(
	senderAddress string,
	senderName string,
	recipientAddresses []string,
	hostName string,
	hostPort int,
	userName string,
	password string) (writer *smtpWriter, err error) {
	return &smtpWriter{
		smtp.PlainAuth("", userName, password, hostName),
		fmt.Sprintf("%s:%d", hostName, hostPort),
		senderAddress,
		senderName,
		recipientAddresses,
	}, nil
}

func prepareMessage(senderAddr, senderName, subject string, body []byte) []byte {
	// Composed according to RFC 5321
	pattern := "From: %s <%s>\nSubject: %s\n"
	h := []byte(fmt.Sprintf(pattern, senderName, senderAddr, subject))
	return append(h, body...)
}

// Write pushes a text message properly composed according to RFC 5321
// to a post server, which sends it to the recipients
func (smtpw *smtpWriter) Write(data []byte) (int, error) {
	err := smtp.SendMail(
		smtpw.hostNameWithPort,
		smtpw.auth,
		smtpw.senderAddress,
		smtpw.recipientAddresses,
		prepareMessage(smtpw.senderAddress, smtpw.senderName, subjectPhrase, data),
	)
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

// Close closes down SMTP-connection 
func (smtpWriter *smtpWriter) Close() error {
	// Do nothing as Write method opens and closes connection automatically
	return nil
}