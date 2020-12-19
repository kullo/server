/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"net/http"

	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/validation"
	"github.com/emicklei/go-restful"
)

type keysSymmWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.KeysSymm
}

func NewKeysSymm() *keysSymmWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/keys/symm").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.KeysSymm{}
	webservice := &keysSymmWebservice{RestfulWebService: service, dao: model}

	// private (filtered)
	service.Route(service.GET("").To(webservice.getEntry))
	service.Route(service.PUT("").To(webservice.insertOrUpdateEntry))

	service.Filter(AuthFilter)
	return webservice
}

func (ws *keysSymmWebservice) readEntryFromBody(request *restful.Request) (*dao.KeysSymmEntry, error) {
	entry := &dao.KeysSymmEntry{}
	err := request.ReadEntity(entry)
	if err != nil {
		return nil, ErrBadRequestBodyFormat
	}

	// validation
	entry.LoginKey, err = validation.ValidateLoginKey(entry.LoginKey)
	if err != nil {
		return nil, validation.NewValidationError("loginKey", err)
	}
	entry.PrivateDataKey, err = validation.ValidatePrivateDataKey(entry.PrivateDataKey)
	if err != nil {
		return nil, validation.NewValidationError("privateDataKey", err)
	}

	return entry, nil
}

func (ws *keysSymmWebservice) insertOrUpdateEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")

	entry, err := ws.readEntryFromBody(request)
	if err != nil {
		writeRequestValidationError(response, err)
		return
	}

	// store loginKey only as a hash
	hash := sha512.Sum512([]byte(entry.LoginKey))
	entry.LoginKey = base64.StdEncoding.EncodeToString(hash[:])

	err = ws.dao.InsertOrUpdateEntry(address, entry)
	if err != nil {
		writeServerError(err, response)
		return
	}

	writeEmptyJson(response, http.StatusOK)
}

func (ws *keysSymmWebservice) getEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")

	entry, err := ws.dao.GetEntry(address)
	switch {
	case err == sql.ErrNoRows:
		// return empty entry if none is stored yet
		entry = &dao.KeysSymmEntry{}
	case err != nil:
		writeServerError(err, response)
		return
	}

	entry.LoginKey = ""
	response.WriteEntity(entry)
}
