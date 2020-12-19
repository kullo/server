/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package notifications

import (
	"fmt"
	"strconv"

	"github.com/tiwoc/gcm"
)

// also used for Android v27
func sendIosNotification(sender *GcmSender, notification *PushNotification, tokens []string) error {
	gcmMessage := new(gcm.Message)
	gcmMessage.RegistrationIDs = tokens
	gcmMessage.ContentAvailable = true
	gcmMessage.Data = make(map[string]interface{})

	// make it through iOS even when app has been force-closed
	gcmMessage.Priority = "high"

	switch notification.Type {
	case PushTypeIncomingMessage:
		gcmNotification := new(gcm.Notification)

		// only visible on Android and Apple Watch
		gcmNotification.Title = "Kullo"                                // iOS: only <8.2
		gcmNotification.TitleLocKey = "notification_title_new_message" // iOS: only >=8.2

		// used on all platforms
		gcmNotification.Body = "You received a new Kullo message"    // iOS: only <8.2
		gcmNotification.BodyLocKey = "notification_body_new_message" // iOS: only >=8.2
		gcmNotification.Sound = "default"

		// only used on iOS
		if notification.UnreadMessages >= 0 {
			gcmNotification.Badge = strconv.Itoa(notification.UnreadMessages)
		}
		if notification.MessageId >= 0 {
			gcmMessage.Data["messageId"] = notification.MessageId
		}

		// only used on Android
		gcmNotification.ClickAction = "net.kullo.action.SYNC"
		gcmNotification.Icon = "kullo_notification"

		gcmMessage.Notification = gcmNotification
		gcmMessage.Data["action"] = "new_message"
		gcmMessage.CollapseKey = "new_message"

	case PushTypeOther:
		gcmMessage.Data["action"] = "other"
		gcmMessage.CollapseKey = "other"

	default:
		return fmt.Errorf("Unknown push type: %s", notification.Type)
	}

	return sendGcmMessage(gcmMessage, sender, notification.Address)
}
