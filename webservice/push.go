/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"net/http"

	"bitbucket.org/kullo/server/dao"
	"github.com/emicklei/go-restful"
)

type pushWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.NotificationsGcm
}

func NewPush() *pushWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/push").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.NotificationsGcm{}
	webservice := &pushWebservice{RestfulWebService: service, dao: model}

	// private (filtered)
	service.Route(service.POST("/gcm").To(webservice.postGcm))
	// This does not support slashes in the token string!
	service.Route(service.DELETE("/gcm/{token}").To(webservice.deleteGcm))

	service.Filter(AuthFilter)
	return webservice
}

func gcmEntryIsValid(entry *dao.NotificationsGcmEntry) bool {
	if entry.RegistrationToken == "" {
		return false
	}
	if !(entry.Environment == "android" || entry.Environment == "ios") {
		return false
	}
	return true
}

func (ws *pushWebservice) readGcmEntryFromBody(request *restful.Request, response *restful.Response) (*dao.NotificationsGcmEntry, bool) {
	entry := &dao.NotificationsGcmEntry{}
	err := request.ReadEntity(entry)
	if err != nil || !gcmEntryIsValid(entry) {
		writeClientError(response, http.StatusBadRequest, "error in request body format")
		return nil, false
	}

	return entry, true
}

func (ws *pushWebservice) postGcm(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")

	entry, ok := ws.readGcmEntryFromBody(request, response)
	if !ok {
		return
	}

	err := ws.dao.InsertEntry(address, entry)
	if err != nil {
		writeServerError(err, response)
		return
	}

	writeEmptyJson(response, http.StatusOK)
}

func (ws *pushWebservice) deleteGcm(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	token := request.PathParameter("token")

	rowsDeleted, err := ws.dao.DeleteEntry(address, token)
	if err != nil {
		writeServerError(err, response)
		return
	}

	if rowsDeleted > 0 {
		writeEmptyJson(response, http.StatusOK)
	} else {
		writeEmptyJson(response, http.StatusNotFound)
	}
}
