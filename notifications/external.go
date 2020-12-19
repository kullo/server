/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package notifications

import (
	"fmt"
	"log"
	"os/exec"
)

const programDir = "/opt/kulloserver/config/hooks/"

type messageType int

const (
	msgTypeWelcome messageType = iota
	msgTypeReset
	msgTypeMessageNotification
)

type messageQueueEntry struct {
	msgType messageType
	data    interface{}
}

type welcomeData struct {
	address  string
	language string
}

type resetData struct {
	address  string
	language string
}

type messageNotificationData struct {
	kulloAddress string
	emailAddress string
	username     string
	cancelSecret string
	language     string
}

var messageQueue = make(chan messageQueueEntry, 100)

func startWorkerForExternalMessages() {
	go func() {
		for {
			entry := <-messageQueue

			var program string
			var args []string

			switch entry.msgType {
			case msgTypeWelcome:
				program = "welcome"
				data := entry.data.(*welcomeData)
				args = append(args, data.address)
				args = append(args, data.language)
			case msgTypeReset:
				program = "reset"
				data := entry.data.(*resetData)
				args = append(args, data.address)
				args = append(args, data.language)
			case msgTypeMessageNotification:
				program = "message_notification"
				data := entry.data.(*messageNotificationData)
				args = append(args, data.kulloAddress)
				args = append(args, data.emailAddress)
				args = append(args, data.username)
				args = append(args, data.cancelSecret)
				args = append(args, data.language)
			default:
				log.Print(fmt.Sprintf(
					"[notifications] unknown msgType: '%d'",
					entry.msgType))
				continue
			}

			programPath := programDir + program
			cmd := exec.Command(programPath, args...)
			err := cmd.Run()
			if err != nil {
				log.Print(fmt.Sprintf(
					"Running the '%s' program failed. Args: %+v",
					program, args))
			}
		}
	}()
}

func SendWelcomeMessage(address string, language string) {
	messageQueue <- messageQueueEntry{
		msgType: msgTypeWelcome,
		data: &welcomeData{
			address:  address,
			language: language,
		}}
}

func SendResetMessage(address string, language string) {
	messageQueue <- messageQueueEntry{
		msgType: msgTypeReset,
		data: &resetData{
			address:  address,
			language: language,
		}}
}

func SendMessageNotification(kulloAddress string, emailAddress string,
	username string, cancelSecret string, language string) {

	messageQueue <- messageQueueEntry{
		msgType: msgTypeMessageNotification,
		data: &messageNotificationData{
			kulloAddress: kulloAddress,
			emailAddress: emailAddress,
			username:     username,
			cancelSecret: cancelSecret,
			language:     language,
		}}
}
