/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"fmt"

	"bitbucket.org/kullo/server/dao"
	"github.com/emicklei/go-restful"
)

type accountInfo struct {
	SettingsLocation string `json:"settingsLocation"`
	PlanName         string `json:"planName"`
	StorageQuota     uint64 `json:"storageQuota"`
	StorageUsed      uint64 `json:"storageUsed"`
}

type accountWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.Users
	messagesDao       *dao.Messages
}

func NewAccount() *accountWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/account").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.Users{}
	messagesModel := &dao.Messages{}
	webservice := &accountWebservice{RestfulWebService: service, dao: model, messagesDao: messagesModel}

	// private (filtered)
	service.Route(service.GET("/info").To(webservice.getInfo))

	service.Filter(AuthFilter)
	return webservice
}

func (ws *accountWebservice) getInfo(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")

	entry, err := ws.dao.GetEntry(address)
	if err != nil {
		writeServerError(err, response)
		return
	}

	if entry.WebloginUsername == "" || entry.WebloginSecret == "" {
		response.WriteEntity(make(map[string]string))
		return
	}

	storageUsed, err := ws.messagesDao.GetStorageSize(address)
	if err != nil {
		writeServerError(err, response)
		return
	}

	info := &accountInfo{}
	info.SettingsLocation =
		fmt.Sprintf("https://accounts.kullo.net/login/?u=%s&s=%s",
			entry.WebloginUsername, entry.WebloginSecret)
	info.PlanName = entry.PlanName
	info.StorageQuota = entry.StorageQuota
	info.StorageUsed = storageUsed
	response.WriteEntity(info)
}
