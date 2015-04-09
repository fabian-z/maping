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

package main

import (
	"github.com/nfdesign/maping/imapclient"
	"github.com/nfdesign/maping/smtpclient"
	"log"
	"strconv"
)

type result struct {
	rx, tx                 int64
	slTx, ilTx, slRx, ilRx []byte
}

func ping(workerid int, emailaccA *emailAccount, emailaccB *emailAccount) *result {
	var testbody = config.testBody
	var logprefix = "Worker " + strconv.Itoa(workerid) + ": "

	var (
		sTx, iTx, tx, sRx, iRx, rx int64
		slTx, ilTx, slRx, ilRx     []byte
		subjectTx, subjectRx       string
		err                        error
	)
	iSettings := &imapclient.ImapSettings{config.imapSettings.loadRecent, config.imapSettings.timeout, config.imapSettings.timeoutRcv, config.imapSettings.waitTime}

	//hacky
	for {
		log.Println(logprefix + "TX")

		subjectTx = "[maping]" + GetRandomString(15)
		log.Println(logprefix + "Sending mail from " + emailaccA.smtpServer + " to " + emailaccB.smtpServer)

		sTx, slTx, err = smtpclient.Send(emailaccA.smtpServer, emailaccA.username, emailaccA.password, emailaccA.explicitSSLSMTP, emailaccA.username, emailaccB.username, subjectTx, testbody)
		log.Println(logprefix + "Checking for mail from " + emailaccA.smtpServer + " on " + emailaccB.imapServer)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		iTx, ilTx, err = imapclient.ConnectAndCheck(emailaccB.imapServer, emailaccB.username, emailaccB.password, emailaccB.explicitSSLIMAP, subjectTx, iSettings)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		tx = iTx - sTx

		if iTx <= 0 {
			//some error likely occured
			tx = -1
		}
		break

	}

	for {

		log.Println(logprefix + "RX")
		subjectRx = "[maping]" + GetRandomString(15)

		log.Println(logprefix + "Sending mail from " + emailaccB.smtpServer + " to " + emailaccA.smtpServer)

		sRx, slRx, err = smtpclient.Send(emailaccB.smtpServer, emailaccB.username, emailaccB.password, emailaccB.explicitSSLSMTP, emailaccB.username, emailaccA.username, subjectRx, testbody)

		if err != nil {
			log.Print(logprefix + err.Error())
			tx = -1
			break
		}

		log.Println(logprefix + "Checking for mail from " + emailaccB.smtpServer + " on " + emailaccA.imapServer)
		iRx, ilRx, err = imapclient.ConnectAndCheck(emailaccA.imapServer, emailaccA.username, emailaccA.password, emailaccA.explicitSSLIMAP, subjectRx, iSettings)

		if err != nil {
			log.Print(logprefix + err.Error())
			rx = -1
			break
		}

		log.Println(logprefix + "Received mail from " + emailaccB.smtpServer + " on " + emailaccA.imapServer)

		rx = iRx - sRx

		if rx <= 0 {
			//some error likely occured
			rx = -1
		}
		break
	}

	return &result{rx, tx, slTx, ilTx, slRx, ilRx}

}
