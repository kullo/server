/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package logging

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/emicklei/go-restful"
)

var errorLogHandle *os.File
var accessLogger *log.Logger
var accessLogHandle *os.File

func OpenErrorLog(filename string) {
	if filename == "" {
		errorLogHandle = nil
		return
	}

	oldHandle := errorLogHandle

	var err error
	errorLogHandle, err = os.OpenFile(filename,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0640)
	if err != nil {
		log.Fatal("Couldn't open error log: " + err.Error())
	}
	log.SetOutput(errorLogHandle)

	if oldHandle != nil {
		oldHandle.Close()
	}
}

func CloseErrorLog() {
	if errorLogHandle != nil {
		errorLogHandle.Close()
	}
}

// Remove Kullo address from the first path component, if present.
// Also leave out query, as it could include potentially unique
// modification dates.
func censor(reqestUrl *url.URL) string {
	path := reqestUrl.Path
	parts := strings.SplitN(path, "/", 3)
	if len(parts) > 1 && strings.ContainsRune(parts[1], '#') {
		parts[1] = "<address>"
		path = strings.Join(parts, "/")
	}
	return path
}

func OpenAccessLog(filename string) {
	if filename == "" {
		accessLogHandle = nil
		accessLogger = nil
		return
	}

	oldHandle := accessLogHandle

	var err error
	accessLogHandle, err = os.OpenFile(filename,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0640)
	if err != nil {
		log.Fatal("Couldn't open access log: " + err.Error())
	}
	accessLogger = log.New(accessLogHandle, "", log.Ldate|log.Ltime)

	if oldHandle != nil {
		oldHandle.Close()
	}
}

func CloseAccessLog() {
	accessLogger = nil
	if accessLogHandle != nil {
		accessLogHandle.Close()
	}
}

func AccessLoggingFilter() restful.FilterFunction {
	return func(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
		startingTime := time.Now()
		chain.ProcessFilter(req, resp)
		elapsedTime := time.Since(startingTime)

		if accessLogger != nil {
			accessLogger.Printf("%s %s %d %.6fs %db",
				req.Request.Method,
				censor(req.Request.URL),
				resp.StatusCode(),
				elapsedTime.Seconds(),
				resp.ContentLength())
		}
	}
}
