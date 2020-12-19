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

	"github.com/tiwoc/gcm"
)

func sendAndroidNotification(sender *GcmSender, notification *PushNotification, tokens []string) error {
	gcmMessage := new(gcm.Message)
	gcmMessage.RegistrationIDs = tokens
	gcmMessage.Data = make(map[string]interface{})

	switch notification.Type {
	case PushTypeIncomingMessage:
		gcmMessage.Data["action"] = "new_message"
		gcmMessage.CollapseKey = "new_message"
		gcmMessage.Priority = "high"

	case PushTypeOther:
		gcmMessage.Data["action"] = "other"
		gcmMessage.CollapseKey = "other"

	default:
		return fmt.Errorf("Unknown push type: %s", notification.Type)
	}

	if notification.UnreadMessages >= 0 {
		gcmMessage.Data["badge"] = notification.UnreadMessages
	}
	if notification.MessageId >= 0 {
		gcmMessage.Data["messageId"] = notification.MessageId
	}

	return sendGcmMessage(gcmMessage, sender, notification.Address)
}

func sendGcmMessage(gcmMessage *gcm.Message, sender *GcmSender, address string) error {
	response, err := sender.Sender.Send(gcmMessage, sender.RetryCount)
	if err != nil {
		return err
	}
	log.Printf("Notifications: %d successes, %d failures",
		response.Success, response.Failure)

	tokens := gcmMessage.RegistrationIDs
	for index, result := range response.Results {
		if len(result.Error) > 0 {
			switch result.Error {

			case "NotRegistered":
				log.Printf("Deleting unregistered GCM token: %s", tokens[index])
				_, err = gcmDao.DeleteEntry(address, tokens[index])
				if err != nil {
					return err
				}

			default:
				log.Printf("[GCM error] token: %s, error: %s",
					tokens[index], result.Error)
			}
		}
		if len(result.RegistrationID) > 0 {
			log.Printf("Updating changed GCM token: %s -> %s",
				tokens[index], result.RegistrationID)
			gcmDao.UpdateToken(address, tokens[index], result.RegistrationID)
		}
	}

	return nil
}
