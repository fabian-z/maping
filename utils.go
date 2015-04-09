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

//Utility functions

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"log"
	"strings"
)

// GetRandomString gets random data of size bytes as base64 encoded string.
func GetRandomString(size int) string {

	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		log.Println(err)
	}

	s := strings.Trim(base64.StdEncoding.EncodeToString(b), "=")
	return s
}

//GzipByteSlice takes the given byte slice and returns a gzip compressed slice.
func GzipByteSlice(b []byte) []byte {

	var z bytes.Buffer
	w := gzip.NewWriter(&z)
	w.Write(b)
	w.Close()
	return z.Bytes()
}
