//LICENSE
//Copyright (c) 2011 The Go Authors. All rights reserved.

//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are
//met:

//* Redistributions of source code must retain the above copyright
//notice, this list of conditions and the following disclaimer.
//* Redistributions in binary form must reproduce the above
//copyright notice, this list of conditions and the following disclaimer
//in the documentation and/or other materials provided with the
//distribution.
//* Neither the name of Google Inc. nor the names of its
//contributors may be used to endorse or promote products derived from
//this software without specific prior written permission.

//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
//"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
//LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
//A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
//OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
//SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
//LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
//DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
//THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
//OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Slightly modified for use in maping

package imapclient

import (
	"errors"
	"time"
)

//Date parser, derived from Go source src/net/mail/message.go
func initDatelayout() {
	// Generate layouts based on RFC 5322, section 3.3.

	dows := [...]string{"", "Mon, "}   // day-of-week
	days := [...]string{"2", "02"}     // day = 1*2DIGIT
	years := [...]string{"2006", "06"} // year = 4*DIGIT / 2*DIGIT
	seconds := [...]string{":05", ""}  // second
	// "-0700 (MST)" is not in RFC 5322, but is common.
	zones := [...]string{"-0700", "MST", "-0700 (MST)"} // zone = (("+" / "-") 4DIGIT) / "GMT" / ...

	for _, dow := range dows {
		for _, day := range days {
			for _, year := range years {
				for _, second := range seconds {
					for _, zone := range zones {
						s := dow + day + " Jan " + year + " 15:04" + second + " " + zone
						dateLayouts = append(dateLayouts, s)
					}
				}
			}
		}
	}
}

func parseDate(date string) (time.Time, error) {

	if len(dateLayouts) == 0 {
		initDatelayout()
	}

	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, date)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("Time from Received header could not be parsed")
}
