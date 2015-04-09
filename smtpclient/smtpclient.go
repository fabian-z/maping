//Copyright 2015 NF Design UG (haftungsbeschraenkt)
//All right reserved.

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at

//  http://www.apache.org/licenses/LICENSE-2.0

//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

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
