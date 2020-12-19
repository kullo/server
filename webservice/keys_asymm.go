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
	"time"

	"bitbucket.org/kullo/server/dao"
	"github.com/emicklei/go-restful"
)

type keysAsymmWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.KeysAsymm
}

func NewKeysAsymm() *keysAsymmWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/keys").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.KeysAsymm{}
	webservice := &keysAsymmWebservice{RestfulWebService: service, dao: model}

	// private (filtered)
	service.Route(service.GET("/private").Filter(AuthFilter).To(webservice.listEntries))
	//service.Route(service.POST("/private").Filter(AuthFilter).To(webservice.createEntry))
	service.Route(service.GET("/private/{id}").Filter(AuthFilter).To(webservice.getEntry))
	//service.Route(service.PATCH("/private/{id}").Filter(AuthFilter).To(webservice.revokeEntry))

	// public (unfiltered)
	service.Route(service.GET("/public").Filter(UserFilter).To(webservice.listPublicEntries))
	service.Route(service.GET("/public/{id}").Filter(UserFilter).To(webservice.getPublicEntry))

	return webservice
}

type entryFilter func(entry *dao.KeysAsymmEntry) interface{}

func nopFilter(entry *dao.KeysAsymmEntry) interface{} {
	return entry
}

func privKeyFilter(entry *dao.KeysAsymmEntry) interface{} {
	return entry.PubkeyEntry
}

func (ws *keysAsymmWebservice) readEntryFromBody(request *restful.Request, response *restful.Response) (*dao.KeysAsymmEntry, bool) {
	entry := &dao.KeysAsymmEntry{}
	err := request.ReadEntity(entry)
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "error in request body format")
		return nil, false
	}
	return entry, true
}

func (ws *keysAsymmWebservice) listEntriesHelper(request *restful.Request, response *restful.Response, filter entryFilter) {
	address := request.PathParameter("address")
	//FIXME get date and type if present
	keyType := ""
	date := time.Time{}

	rows, err := ws.dao.GetList(address, keyType, date)
	if err != nil {
		writeServerError(err, response)
		return
	}
	defer rows.Close()

	entries := make([]interface{}, 0, 100)
	for rows.Next() {
		entry, err := ws.dao.GetNextEntry(rows)
		if err != nil {
			writeServerError(err, response)
			return
		}
		entries = append(entries, filter(entry))
	}
	if rows.Err() != nil {
		writeServerError(rows.Err(), response)
		return
	}

	response.WriteEntity(&entries)
}

func (ws *keysAsymmWebservice) listEntries(request *restful.Request, response *restful.Response) {
	ws.listEntriesHelper(request, response, nopFilter)
}

func (ws *keysAsymmWebservice) listPublicEntries(request *restful.Request, response *restful.Response) {
	ws.listEntriesHelper(request, response, privKeyFilter)
}

func (ws *keysAsymmWebservice) createEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")

	entry, ok := ws.readEntryFromBody(request, response)
	if !ok {
		return
	}

	entry.ID = 0
	//FIXME validate type and valid dates
	entry.ValidFrom = time.Now().Add(-15 * time.Minute)
	entry.ValidUntil = time.Now().Add(3650 * 24 * time.Hour)
	entry.Revocation = ""

	id, err := ws.dao.InsertEntry(address, entry)
	if err != nil {
		writeServerError(err, response)
		return
	}

	response.WriteEntity(id)
}

func getIdOrLatest(request *restful.Request, response *restful.Response) (bool, uint32, bool) {
	idStr := request.PathParameter("id")
	if idStr == "latest-enc" {
		return true, 0, true
	}
	id, ok := getID(request, response)
	return false, id, ok
}

func (ws *keysAsymmWebservice) getEntryHelper(request *restful.Request, response *restful.Response, filter entryFilter) {
	address := request.PathParameter("address")
	latest, id, ok := getIdOrLatest(request, response)
	if !ok {
		return
	}

	if latest {
		id = dao.LATEST_ENCRYPTION_PUBKEY
	}
	entry, err := ws.dao.GetEntry(address, id)
	switch {
	case err == sql.ErrNoRows:
		writeClientError(response, http.StatusNotFound, "entry not found")
		return
	case err != nil:
		writeServerError(err, response)
		return
	}

	response.WriteEntity(filter(entry))
}

func (ws *keysAsymmWebservice) getEntry(request *restful.Request, response *restful.Response) {
	ws.getEntryHelper(request, response, nopFilter)
}

func (ws *keysAsymmWebservice) getPublicEntry(request *restful.Request, response *restful.Response) {
	ws.getEntryHelper(request, response, privKeyFilter)
}

func (ws *keysAsymmWebservice) revokeEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	id, ok := getID(request, response)
	if !ok {
		return
	}

	entry, ok := ws.readEntryFromBody(request, response)
	//FIXME validate revocation

	err := ws.dao.SetRevocation(address, id, entry.Revocation)
	writeEntityOrModificationErr(make(map[string]string), err, response)
}
