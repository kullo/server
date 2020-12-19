/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"database/sql"
	"net/http"

	"bitbucket.org/kullo/server/dao"
	"github.com/emicklei/go-restful"
)

type profileWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.Profile
}

func NewProfile() *profileWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/profile").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.Profile{}
	webservice := &profileWebservice{RestfulWebService: service, dao: model}

	// private (filtered)
	service.Route(service.GET("").To(webservice.listEntries))
	service.Route(service.GET("/{key}").To(webservice.getEntry))
	service.Route(service.PUT("/{key}").To(webservice.modifyEntry))

	service.Filter(AuthFilter)
	return webservice
}

func (ws *profileWebservice) readEntryFromBody(request *restful.Request, response *restful.Response, permissive bool) *dao.ProfileEntry {
	entry := &dao.ProfileEntry{}
	err := request.ReadEntity(entry)
	if err != nil || (!permissive && !entry.IsValid()) {
		writeClientError(response, http.StatusBadRequest, "error in request body format")
		return nil
	}
	return entry
}

func (ws *profileWebservice) listEntries(request *restful.Request, response *restful.Response) {
	address, modifiedAfter, _, ok := getListParameters(request, response)
	if !ok {
		return
	}

	result := &dataWithResultCount{}
	var rows *sql.Rows
	var err error
	rows, err = ws.dao.GetList(address, modifiedAfter)
	if err != nil {
		writeServerError(err, response)
		return
	}
	defer rows.Close()

	result.Data = make([]interface{}, 0, 10)
	for rows.Next() {
		entry, err := ws.dao.GetNextEntry(rows)
		if err != nil {
			writeServerError(err, response)
			return
		}
		result.Data = append(result.Data, entry)
	}
	if rows.Err() != nil {
		writeServerError(rows.Err(), response)
		return
	}

	result.ResultsReturned = uint32(len(result.Data))
	result.ResultsTotal = result.ResultsReturned
	response.WriteEntity(&result)
}

func (ws *profileWebservice) getEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	key := request.PathParameter("key")

	entry, err := ws.dao.GetEntry(address, key)

	switch {
	case err == sql.ErrNoRows:
		writeClientError(response, http.StatusNotFound, "entry not found")
		return
	case err != nil:
		writeServerError(err, response)
		return
	}

	response.WriteEntity(entry)
}

func (ws *profileWebservice) modifyEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	key := request.PathParameter("key")
	lastModified, ok := getLastModified(request, response)
	if !ok {
		return
	}

	entry := ws.readEntryFromBody(request, response, true)
	if entry == nil {
		return
	}
	entry.SetKeyLastModified(key, lastModified)
	if !entry.IsValid() {
		writeClientError(response, http.StatusBadRequest, "invalid entry")
		return
	}

	meta, err := ws.dao.ModifyEntry(address, entry)
	writeEntityOrModificationErr(meta, err, response)
}
