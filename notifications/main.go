/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package notifications

func StartWorkers(gcmApiKey string) {
	startWorkerForInternalMessages(gcmApiKey)
	startWorkerForExternalMessages()
}
