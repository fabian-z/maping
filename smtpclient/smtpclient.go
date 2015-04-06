// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

// smtpclient
package smtpclient

// Using modified net/smtp package to provide plain auth over SSL/TLS
// and CRAM-MD5 detection with plain fallback over SendMail func

import (
	"github.com/nfdesign/maping/smtpclient/smtpssl"
	"bytes"
	"log"
	"text/template"
	"time"
)

type EmailTemplateData struct {
	From    string
	To      string
	Date    string
	Subject string
	Body    string
}

const explicitssl_port string = "465"
const smtp_mail_submission_port string = "587"

const emailTemplate = `From: {{.From}}
To: {{.To}}
Date: {{.Date}}
Subject: {{.Subject}}

{{.Body}}
`

func Send(host string, user string, password string, explicitssl bool, fromaddr string, toaddr string, subject string, body string) (int64, []byte, error) {

	var (
		err error
		buf bytes.Buffer
	)

	date := time.Now()

	context := &EmailTemplateData{fromaddr,
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
			host+":"+explicitssl_port,
			plainauth,
			cramauth,
			context.From,
			[]string{context.To},
			buf.Bytes(),
		)
		if err != nil {
			log.Print(err)
			return -1, nil, err
		}
		return date.Unix(), sl, nil
	} else {

		sl, err := smtpssl.SendMail(
			host+":"+smtp_mail_submission_port,
			plainauth,
			cramauth,
			context.From,
			[]string{context.To},
			buf.Bytes(),
		)
		if err != nil {
			log.Print(err)
			return -1, nil, err
		}
		return date.Unix(), sl, nil
	}

}
