// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

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
