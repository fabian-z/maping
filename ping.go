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

type result struct {
	rx, tx                     int64
	sl_tx, il_tx, sl_rx, il_rx []byte
}

func ping(workerid int, emailacc_a *emailAccount, emailacc_b *emailAccount) *result {
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
		log.Println(logprefix + "Sending mail from " + emailacc_a.smtpServer + " to " + emailacc_b.smtpServer)

		s_tx, sl_tx, err = smtpclient.Send(emailacc_a.smtpServer, emailacc_a.username, emailacc_a.password, emailacc_a.explicitSSLSMTP, emailacc_a.username, emailacc_b.username, subject_tx, testbody)
		log.Println(logprefix + "Checking for mail from " + emailacc_a.smtpServer + " on " + emailacc_b.imapServer)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		i_tx, il_tx, err = imapclient.ConnectAndCheck(emailacc_b.imapServer, emailacc_b.username, emailacc_b.password, emailacc_b.explicitSSLIMAP, subject_tx, imap_settings)

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

		log.Println(logprefix + "Sending mail from " + emailacc_b.smtpServer + " to " + emailacc_a.smtpServer)

		s_rx, sl_rx, err = smtpclient.Send(emailacc_b.smtpServer, emailacc_b.username, emailacc_b.password, emailacc_b.explicitSSLSMTP, emailacc_b.username, emailacc_a.username, subject_rx, testbody)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		log.Println(logprefix + "Checking for mail from " + emailacc_b.smtpServer + " on " + emailacc_a.imapServer)
		i_rx, il_rx, err = imapclient.ConnectAndCheck(emailacc_a.imapServer, emailacc_a.username, emailacc_a.password, emailacc_a.explicitSSLIMAP, subject_rx, imap_settings)

		if err != nil {
			log.Print(logprefix + err.Error())
			rx = -1
			break
		}

		log.Println(logprefix + "Received mail from " + emailacc_b.smtpServer + " on " + emailacc_a.imapServer)

		rx = i_rx - s_rx

		if rx <= 0 {
			//some error likely occured
			rx = -1
		}
		break
	}

	return &result{rx, tx, sl_tx, il_tx, sl_rx, il_rx}

}
