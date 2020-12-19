/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"golang.org/x/text/language"

	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/dbconn"
	"bitbucket.org/kullo/server/logging"
	"bitbucket.org/kullo/server/notifications"
	"bitbucket.org/kullo/server/util"
	"bitbucket.org/kullo/server/webservice"
	"github.com/emicklei/go-restful"
	"github.com/kylelemons/go-gypsy/yaml"
	_ "github.com/lib/pq"
)

func openDb(dbEnvironment string, configDir string) {
	conf, err := yaml.ReadFile(configDir + "/dbconf.yml")
	if err != nil {
		log.Fatal(err)
	}

	dbstr, err := conf.Get(dbEnvironment + ".open")
	if err != nil {
		log.Fatal(err)
	}

	err = dbconn.Open("postgres", dbstr)
	if err != nil {
		panic(err)
	}
}

func statusHandler(rw http.ResponseWriter, req *http.Request) {
	users := dao.Users{}
	_, err := users.UserExists("hi#kullo.net")
	if err != nil {
		util.LogServerError(err)
		rw.WriteHeader(http.StatusInternalServerError)
		io.WriteString(rw, "HTTP 500\n\nCouldn't access database")
		return
	}

	rw.WriteHeader(http.StatusOK)
	io.WriteString(rw, "HTTP 200\n\nUp and running")
}

func main() {
	// configure logging (should be the first thing to be executed)
	log.SetFlags(log.Flags() | log.Lshortfile)

	// command line parsing
	port := flag.Int("port", 8001, "Server port")
	configDir := flag.String("configDir", "./config", "configuration directory")
	dbEnvironment := flag.String("env", "local", "database configuration environment name")
	domain := flag.String("domain", "kullo.test", "domain part of this server's addresses")
	gcmApiKey := flag.String("gcmApiKey", "", "API key for Google Cloud Messaging")
	accessLogFile := flag.String("accessLogFile", "/tmp/kulloserver-access.log", "file name of the access log")
	errorLogFile := flag.String("errorLogFile", "", "file name of the error log")
	cpuProfile := flag.String("cpuprofile", "", "write cpu profile to given file")
	memProfile := flag.String("memprofile", "", "write memory profile to given file")
	flag.Parse()

	logging.OpenErrorLog(*errorLogFile)
	defer logging.CloseErrorLog()
	logging.OpenAccessLog(*accessLogFile)
	defer logging.CloseAccessLog()

	// CPU profiling
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// handle SIGINT (Ctrl+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-signalChan
			switch sig {

			case syscall.SIGINT:
				// write memory profile
				if *memProfile != "" {
					f, err := os.Create(*memProfile)
					if err != nil {
						log.Fatal(err)
					}
					pprof.WriteHeapProfile(f)
					f.Close()
				}

				os.Exit(0)

			case syscall.SIGUSR1:
				// reopen logs
				log.Println("Reopening logs")
				logging.OpenErrorLog(*errorLogFile)
				logging.OpenAccessLog(*accessLogFile)
			}
		}
	}()

	// status page
	http.HandleFunc("/status", statusHandler)

	// open DB
	openDb(*dbEnvironment, *configDir)
	defer dbconn.Close()

	webservice.SetAvailableLanguages(language.English, language.German)

	// set up restful
	restful.Filter(logging.AccessLoggingFilter())
	restful.Filter(webservice.LanguageFilter)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	restful.PrettyPrintResponses = false
	restful.Add(webservice.NewAccounts(*domain).RestfulWebService)
	restful.Add(webservice.NewAccount().RestfulWebService)
	restful.Add(webservice.NewMessages().RestfulWebService)
	restful.Add(webservice.NewKeysSymm().RestfulWebService)
	restful.Add(webservice.NewKeysAsymm().RestfulWebService)
	restful.Add(webservice.NewPush().RestfulWebService)
	restful.Add(webservice.NewProfile().RestfulWebService)

	notifications.StartWorkers(*gcmApiKey)

	log.Print(fmt.Sprintf("Starting HTTP server for %s on port %d ...", *domain, *port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
