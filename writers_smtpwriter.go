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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"path/filepath"
	"strings"
)

const (
	subjectPhrase = "Diagnostic message from server: "
	// Message subject pattern composed according to RFC 5321.
	rfc5321SubjectPattern = "From: %s <%s>\nSubject: %s\n"
)

// smtpWriter is used to send emails via given SMTP-server.
type smtpWriter struct {
	auth               smtp.Auth
	hostName           string
	hostPort           string
	hostNameWithPort   string
	senderAddress      string
	senderName         string
	recipientAddresses []string
	//caCertDirPaths []string
	config *tls.Config
}

// newSmtpWriter returns a new SMTP writer.
func newSmtpWriter(
	senderAddress, senderName string,
	recipientAddresses []string,
	hostName, hostPort, userName, password string,
	caCertDirPaths []string,
) (writer *smtpWriter, err error) {
	var config *tls.Config
	var e error
	// Define TLS Config iff caCertDirPaths are given.
	if caCertDirPaths != nil && len(caCertDirPaths) > 0 {
		config, e = getTLSConfig(caCertDirPaths, hostName)
		if e != nil {
			writer = nil
			err = e
			return
		}
	}
	writer = &smtpWriter{
		smtp.PlainAuth("", userName, password, hostName),
		hostName,
		hostPort,
		fmt.Sprintf("%s:%s", hostName, hostPort),
		senderAddress,
		senderName,
		recipientAddresses,
		config,
	}
	return
}

func prepareMessage(senderAddr, senderName, subject string, body []byte) []byte {
	h := []byte(fmt.Sprintf(rfc5321SubjectPattern, senderName, senderAddr, subject))
	return append(h, body...)
}

// getTLSConfig gets paths of folders with X.509 CA PEM files,
// host server name and tries to create an appropriate TLS.Config.
func getTLSConfig(pemDirPaths []string, hostName string) (config *tls.Config, err error) {
	var pemFilePaths []string
	for _, pdp := range pemDirPaths {
		filePaths, e := fileSystemWrapper.GetDirFileNames(pdp, true)
		if e != nil {
			return nil, e
		}
		for _, fp := range filePaths {
			if strings.ToUpper(filepath.Ext(fp)) == ".PEM" {
				pemFilePaths = append(pemFilePaths, fp)
			}
		}
	}

	pemEncodedContent := []byte{}
	var (
		e     error
		bytes []byte
	)
	// Put together all the PEM files to decode them as a whole byte slice.
	for _, pfp := range pemFilePaths {
		if bytes, e = ioutil.ReadFile(pfp); e == nil {
			pemEncodedContent = append(pemEncodedContent, bytes...)
		} else {
			err = fmt.Errorf("Cannot read file: %s", pfp)
			return
		}
	}
	config = &tls.Config{RootCAs: x509.NewCertPool(), ServerName: hostName}
	isAppended := config.RootCAs.AppendCertsFromPEM(pemEncodedContent)
	if !isAppended {
		// Extract this into a separate error.
		err = errors.New("Invalid PEM content")
		return
	}
	return
}

// SendMail accepts TLS configuration, connects to the server at addr,
// switches to TLS if possible, authenticates with mechanism a if possible,
// and then sends an email from address from, to addresses to, with message msg.
func sendMailWithTLSConfig(config *tls.Config, addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	// Check if the server supports STARTTLS extension.
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	// Check if the server supports AUTH extension and use given smtp.Auth.
	if a != nil {
		if isSupported, _ := c.Extension("AUTH"); isSupported {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	// Portion of code from the official smtp.SendMail function,
	// see http://golang.org/src/pkg/net/smtp/smtp.go.
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// Write pushes a text message properly composed according to RFC 5321
// to a post server, which sends it to the recipients.
func (smtpw *smtpWriter) Write(data []byte) (int, error) {
	var err error
	msg := prepareMessage(smtpw.senderAddress, smtpw.senderName, subjectPhrase, data)
	if smtpw.config == nil {
		err = smtp.SendMail(
			smtpw.hostNameWithPort,
			smtpw.auth,
			smtpw.senderAddress,
			smtpw.recipientAddresses,
			msg,
		)
	} else {
		err = sendMailWithTLSConfig(
			smtpw.config,
			smtpw.hostNameWithPort,
			smtpw.auth,
			smtpw.senderAddress,
			smtpw.recipientAddresses,
			msg,
		)
	}
	if err != nil {
		return 0, err
	}
	return len(data), nil
}

// Close closes down SMTP connection.
func (smtpWriter *smtpWriter) Close() error {
	// Do nothing as Write method opens and
	// closes connection automatically.
	return nil
}
