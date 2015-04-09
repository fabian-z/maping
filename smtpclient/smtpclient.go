// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package smtpclient

// smtpclient used for testing.
// Using modified net/smtp package to provide plain auth over SSL/TLS
// and CRAM-MD5 detection with plain fallback over SendMail func

import (
	"bytes"
	"github.com/nfdesign/maping/smtpclient/smtpssl"
	"log"
	"text/template"
	"time"
)

type emailTemplateData struct {
	from    string
	to      string
	date    string
	subject string
	body    string
}

const explicitSSLPort string = "465"
const mailSubmissionPort string = "587"

const emailTemplate = `From: {{.From}}
To: {{.To}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Body}}
`

//Send connects to SMTP server with given authentication and connection information. It then sends mail with subject to toaddr and returns
//the unix timestamp when the mail was send as integer, the smtp protocol log as []byte or, if applicable, an non-nil error.
func Send(host string, user string, password string, explicitssl bool, fromaddr string, toaddr string, subject string, body string) (int64, []byte, error) {

	var (
		err error
		buf bytes.Buffer
	)

	date := time.Now()

	context := &emailTemplateData{fromaddr,
		toaddr,
		date.Format(time.RFC1123Z),
		subject,
		body}

	// Create a new template for our SMTP message.
	t := template.New("emailTemplate")
	if t, err = t.Parse(emailTemplate); err != nil {
		log.Print("error trying to parse mail template ", err)
	}

	// Apply values from struct to template
	if err = t.Execute(&buf, context); err != nil {
		log.Print("error trying to execute mail template ", err)
	}

	// Set up authentication information.
	plainauth := smtpssl.PlainAuth(
		"", //identity
		user,
		password,
		host,
	)

	cramauth := smtpssl.CRAMMD5Auth(
		user,
		password,
	)

	if explicitssl {
		sl, err := smtpssl.SendMailSSL(
			host+":"+explicitSSLPort,
			plainauth,
			cramauth,
			context.from,
			[]string{context.to},
			buf.Bytes(),
		)
		if err != nil {
			log.Print(err)
			return -1, nil, err
		}
		return date.Unix(), sl, nil
	}

	sl, err := smtpssl.SendMail(
		host+":"+mailSubmissionPort,
		plainauth,
		cramauth,
		context.from,
		[]string{context.to},
		buf.Bytes(),
	)
	if err != nil {
		log.Print(err)
		return -1, nil, err
	}
	return date.Unix(), sl, nil

}
