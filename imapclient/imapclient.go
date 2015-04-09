// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package imapclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"github.com/mxk/go-imap/imap"
	"log"
	"net/mail"
	"strings"
	"time"
)

var (
	dateLayouts []string
)

const (
	explicitSSLPort string = "993"
)

type byteLogger struct {
	imaplog []byte
}

//ImapSettings defines common timeouts and durations in sec.
type ImapSettings struct {
	LoadRecent int
	Timeout    int
	TimeoutRcv int64
	WaitTime   int
}

func recvUnilateralResponse(unichan chan string) {

	unichan <- "rcv"

}

func fetchAndProcess(c *imap.Client, subject string, loadrecent uint32, unichan chan string) (int64, error) {

	var uid uint32
	var rcvdate time.Time

	expunge := false

	// Open a mailbox with R/W access (synchronous command - no need for imap.Wait)
	c.Select("INBOX", false)
	defer func() {
		//Remember closing the mailbox, expunging as given
		c.Close(expunge)
	}()

	recvUnilateralResponse(unichan)

	// Fetch the headers of the 10 most recent messages
	if c.Mailbox.Messages == 0 {

		return -2, errors.New("No mail found")
	}

	set, _ := imap.NewSeqSet("")
	if c.Mailbox.Messages >= loadrecent {
		set.AddRange(c.Mailbox.Messages-(loadrecent-1), c.Mailbox.Messages)
	} else {
		set.Add("1:*")
	}
	cmd, _ := c.Fetch(set, "UID", "RFC822.HEADER")

	// Process responses while the command is running

	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		c.Recv(-1)
		recvUnilateralResponse(unichan)
		// Process command data
		for _, rsp := range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])

			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {

				//Match even if subject has been wrapped or prefixed/suffixed
				if strings.Contains(msg.Header.Get("Subject"), subject) {

					uid = rsp.MessageInfo().UID

					//Header.Get gets first Received header, which is added by the
					//last SMTP server according to RFC
					//Split by semicolon and use last field to parse date
					arr := strings.Split(msg.Header.Get("Received"), ";")
					//Need some error handling for strings.Split here

					var err error
					rcvdate, err = parseDate(strings.TrimSpace(arr[len(arr)-1]))
					if err != nil {

						log.Println(err)
						return -1, err

					}

				}

			}
		}
		cmd.Data = nil

	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			log.Println("Fetch command aborted")
		} else {
			log.Println("Fetch error:", rsp.Info)
		}
	}

	if (uid != 0 && rcvdate != time.Time{}) {

		//Flag msg as deleted
		del, _ := imap.NewSeqSet("")
		del.AddNum(uid)

		//Waiting for this command introduces errors, e.g. with GMail
		_, err := c.UIDStore(del, "+FLAGS", imap.NewFlagSet(`\Deleted`))

		if err != nil {
			log.Fatal(err)
		}

		recvUnilateralResponse(unichan)

		//Expunge can be expensive. Apply only to our testmail if
		//UIDPLUS is supported

		if _, ok := c.Caps["UIDPLUS"]; ok {

			c.Expunge(del)
			recvUnilateralResponse(unichan)
			expunge = false
		} else {

			//Expunge is applied on c.Close in deferred cleanup
			expunge = true

		}

		return rcvdate.Unix(), nil

	}
	return -2, errors.New("No mail found")

}

func (w *byteLogger) Write(p []byte) (int, error) {

	//This is in conscious violation of the type Writer spec in pkg/io:
	//"Implementations must not retain p."

	w.imaplog = append(w.imaplog, p...)
	return len(p), nil
}

//ConnectAndCheck connects to an IMAP host, waiting for a given subject in INBOX for timeoutrcv seconds.
//It checks for new mail every waittime seconds, but reacts on a unilateral response indicating new mail.
func ConnectAndCheck(host string, user string, password string, explicitssl bool, subject string, config *ImapSettings) (int64, []byte, error) {

	var (
		//Channel for goroutine synchronization
		unichan    = make(chan string, 1)
		force      = make(chan string, 1)
		c          *imap.Client
		rsp        *imap.Response
		timeout    = time.Duration(config.Timeout) * time.Second
		timeoutrcv = time.Duration(config.TimeoutRcv) * time.Second
		waittime   = time.Duration(config.WaitTime) * time.Second
		loadrecent = uint32(config.LoadRecent)
		err        error
	)

	w := &byteLogger{}

	//Log protocol
	imap.DefaultLogger = log.New(w, "", 0)
	imap.DefaultLogMask = imap.LogConn | imap.LogRaw

	if explicitssl == true {

		//Explicit SSL
		conn, err := tls.Dial("tcp", host+":"+explicitSSLPort, nil)

		if err != nil {

			log.Printf("TLS connection to host %s port %s failed", host, explicitSSLPort)
			log.Println(err)
			return -1, nil, err
		}

		// Connect to the server
		c, err = imap.NewClient(conn, host+":"+explicitSSLPort, timeout)
		if err != nil {

			log.Printf("IMAP connection to host %s port %s failed", host, explicitSSLPort)
			log.Println(err)
			return -1, nil, err
		}
	} else {

		// Connect to the server
		c, err = imap.Dial(host)
		if err != nil {

			log.Printf("Plain IMAP connection to server %s failed", host)
			log.Println(err)
			return -1, nil, err
		}
	}

	// Remember to log out and close the connection when finished
	defer func(unichan chan string) {

		c.Logout(timeout)
		recvUnilateralResponse(unichan)
	}(unichan)

	//Discard server greeting (we get this in the protocol log)
	c.Data = nil

	// Enable encryption, if supported by the server
	if c.Caps["STARTTLS"] {

		c.StartTLS(nil)
	}
	recvUnilateralResponse(unichan)
	if c.State() == imap.Login {

		if c.Caps["AUTH=CRAM-MD5"] {

			if _, err = loginCramIMAP(c, user, password); err != nil {

				return -1, nil, errors.New("IMAP Authenticate   username" + user + "error" + err.Error())
			}

		} else {

			if _, err = loginIMAP(c, user, password); err != nil {
				return -1, nil, errors.New("IMAP Login  username" + user + "error" + err.Error())
			}

		}

	}

	// Goroutine to check for new unilateral server data responses if requested on channel

	go func(c *imap.Client, unichan chan string, force chan string) {

		for {
			select {

			case <-unichan:

				if c.State().String() == "Closed" {
					//Stop receiving updates
					return
				}

				for _, rsp = range c.Data {

					w.imaplog = append(w.imaplog, imap.AsBytes(rsp)...)

					if strings.Contains(rsp.String(), "EXISTS") {
						//force check if server gives us a hint on new mail

						force <- "rcv"

					}

				}
				c.Data = nil

			}
		}

	}(c, unichan, force)

	timeoutChannel := time.After(timeoutrcv)
	tick := time.Tick(waittime)

	force <- "rcv"

	// Keep trying until we're timed out, got a result or an error
	for {
		select {

		// Got a timeout! fail with a timeout error
		case <-timeoutChannel:
			return -1, nil, errors.New("Timeout waiting for mail")

		case <-tick:
			// Got a tick

			ret, err := fetchAndProcess(c, subject, loadrecent, unichan)

			if !(ret == -2) {
				//Found no mail, but no hard error occured
				if err != nil {

					return -1, nil, err
				}
				return ret, w.imaplog, nil

			}

		case <-force:
			//Use separate channel to work around first tick delay

			ret, err := fetchAndProcess(c, subject, loadrecent, unichan)

			if !(ret == -2) {
				//Found no mail, but no hard error occured
				if err != nil {

					return -1, nil, err
				}
				return ret, w.imaplog, nil

			}

		}
	}

}
