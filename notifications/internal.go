/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package notifications

import (
	"log"
	"net/http"
	"time"

	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/util"
	"github.com/tiwoc/gcm"
)

var internalMessageQueue = make(chan PushNotification, 1000000)
var gcmApiKey string
var gcmDao = dao.NotificationsGcm{}

type GcmSender struct {
	Sender     gcm.Sender
	RetryCount int
}

func tokensForRegistrationsWithEnv(registrations *[]dao.NotificationsGcmEntry, environment string) []string {
	var result []string
	for _, registration := range *registrations {
		if registration.Environment == environment {
			result = append(result, registration.RegistrationToken)
		}
	}
	return result
}

func startWorkerForInternalMessages(apiKey string) {
	if len(apiKey) == 0 {
		log.Println("No GCM API key is set, can't send push notifications.")
		return
	}
	gcmApiKey = apiKey

	go func() {
		httpClient := new(http.Client)
		// set the timeout for sending a single request (restarted for each retry)
		httpClient.Timeout = 20 * time.Second
		// results in a timeout for sending this notification after approx 2^8s ~ 4min, max 1.5*2^8 ~ 6.5 min
		retryCount := 8
		sender := GcmSender{gcm.Sender{ApiKey: gcmApiKey, Http: httpClient}, retryCount}

		for {
			notification := <-internalMessageQueue
			registrations, err := gcmDao.GetTokens(notification.Address)
			if err != nil {
				util.LogServerError(err)
				continue
			}

			// Android v28+
			tokens := tokensForRegistrationsWithEnv(&registrations, "android")
			if len(tokens) > 0 {
				err = sendAndroidNotification(&sender, &notification, tokens)
				if err != nil {
					util.LogServerError(err)
				}
			}

			// iOS v19+
			tokens = tokensForRegistrationsWithEnv(&registrations, "ios")
			if len(tokens) > 0 {
				err = sendIosNotification(&sender, &notification, tokens)
				if err != nil {
					util.LogServerError(err)
				}
			}
		}
	}()
}

type PushType int

const (
	// An incoming message has been received
	PushTypeIncomingMessage PushType = iota
	// Any other reason for syncing (silent, will not show a notification)
	PushTypeOther
)

type PushNotification struct {
	Type           PushType
	Address        string
	MessageId      int
	UnreadMessages int
}

func SendPushNotifications(notification PushNotification) {
	if len(gcmApiKey) > 0 {
		internalMessageQueue <- notification

		queueLength := len(internalMessageQueue)
		if queueLength >= 10 {
			log.Printf("Internal push notification enqueued; queue length: %d",
				queueLength)
		}
	}
}
