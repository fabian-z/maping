// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

import (
	"github.com/nfdesign/maping/imapclient"
	"github.com/nfdesign/maping/smtpclient"
	"log"
	"strconv"
)

type Result struct {
	RX, TX                     int64
	SL_TX, IL_TX, SL_RX, IL_RX []byte
}

func ping(workerid int, emailacc_a *EmailAccount, emailacc_b *EmailAccount) *Result {
	var testbody string = config.testBody
	var logprefix = "Worker " + strconv.Itoa(workerid) + ": "

	var (
		s_tx, i_tx, tx, s_rx, i_rx, rx int64
		sl_tx, il_tx, sl_rx, il_rx     []byte
		subject_tx, subject_rx         string
		err                            error
	)
	imap_settings := &imapclient.ImapSettings{config.imapSettings.loadRecent, config.imapSettings.timeout, config.imapSettings.timeoutRcv, config.imapSettings.waitTime}

	//hacky
	for {
		log.Println(logprefix + "TX")

		subject_tx = "[maping]" + GetRandomString(15)
		log.Println(logprefix + "Sending mail from " + emailacc_a.SMTPServer + " to " + emailacc_b.SMTPServer)

		s_tx, sl_tx, err = smtpclient.Send(emailacc_a.SMTPServer, emailacc_a.Username, emailacc_a.Password, emailacc_a.ExplicitSSL_SMTP, emailacc_a.Username, emailacc_b.Username, subject_tx, testbody)
		log.Println(logprefix + "Checking for mail from " + emailacc_a.SMTPServer + " on " + emailacc_b.IMAPServer)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		i_tx, il_tx, err = imapclient.ConnectAndCheck(emailacc_b.IMAPServer, emailacc_b.Username, emailacc_b.Password, emailacc_b.ExplicitSSL_IMAP, subject_tx, imap_settings)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		tx = i_tx - s_tx

		if i_tx <= 0 {
			//some error likely occured
			tx = -1
		}
		break

	}

	for {

		log.Println(logprefix + "RX")
		subject_rx = "[maping]" + GetRandomString(15)

		log.Println(logprefix + "Sending mail from " + emailacc_b.SMTPServer + " to " + emailacc_a.SMTPServer)

		s_rx, sl_rx, err = smtpclient.Send(emailacc_b.SMTPServer, emailacc_b.Username, emailacc_b.Password, emailacc_b.ExplicitSSL_SMTP, emailacc_b.Username, emailacc_a.Username, subject_rx, testbody)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		log.Println(logprefix + "Checking for mail from " + emailacc_b.SMTPServer + " on " + emailacc_a.IMAPServer)
		i_rx, il_rx, err = imapclient.ConnectAndCheck(emailacc_a.IMAPServer, emailacc_a.Username, emailacc_a.Password, emailacc_a.ExplicitSSL_IMAP, subject_rx, imap_settings)

		if err != nil {
			log.Print(logprefix + err.Error())
			rx = -1
			break
		}

		log.Println(logprefix + "Received mail from " + emailacc_b.SMTPServer + " on " + emailacc_a.IMAPServer)

		rx = i_rx - s_rx

		if rx <= 0 {
			//some error likely occured
			rx = -1
		}
		break
	}

	return &Result{rx, tx, sl_tx, il_tx, sl_rx, il_rx}

}
