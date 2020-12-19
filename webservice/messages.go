/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/notifications"
	"bitbucket.org/kullo/server/util"
	"github.com/emicklei/go-restful"
)

type createMessageResult struct {
	ID           uint32 `json:"id"`
	LastModified uint64 `json:"lastModified"`
	Received     string `json:"dateReceived"`
}

type messagesWebservice struct {
	RestfulWebService *restful.WebService
	dao               *dao.Messages
}

func NewMessages() *messagesWebservice {
	service := &restful.WebService{}
	service.
		Path("/{address}/messages").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	model := &dao.Messages{}
	webservice := &messagesWebservice{RestfulWebService: service, dao: model}

	// private (filtered)
	service.Route(service.GET("").Filter(AuthFilter).To(webservice.listEntries))
	service.Route(service.GET("/{id}").Filter(AuthFilter).To(webservice.getEntry))
	service.Route(service.PATCH("/{id}").Filter(AuthFilter).To(webservice.modifyMeta))
	service.Route(service.DELETE("/{id}").Filter(AuthFilter).To(webservice.deleteEntry))
	service.Route(service.GET("/{id}/attachments").Filter(AuthFilter).To(webservice.getAttachments))

	// public (unfiltered)
	// JSON body
	service.Route(service.POST("").
		Filter(UserFilter).
		Filter(OptionalAuthFilter).
		To(webservice.createEntryFromJson))
	// multipart body
	service.Route(service.POST("").
		Consumes("multipart/form-data").
		Filter(UserFilter).
		Filter(OptionalAuthFilter).
		To(webservice.createEntryFromMultipart))

	return webservice
}

func (ws *messagesWebservice) readEntryFromJsonBody(request *restful.Request, response *restful.Response) (*dao.MessagesEntry, bool) {
	entry := &dao.MessagesEntry{}
	err := request.ReadEntity(entry)
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "error in request body format")
		return nil, false
	}

	if len(entry.AttachmentsBase64)*3 > dao.MESSAGE_JSON_ATTACHMENTS_MAX_BYTES*4 {
		writeClientError(response, http.StatusBadRequest, "attachments too long")
		return nil, false
	}

	entry.Attachments, err = base64.StdEncoding.DecodeString(entry.AttachmentsBase64)
	entry.AttachmentsBase64 = ""
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "attachments: invalid encoding")
		return nil, false
	}

	return entry, true
}

func (ws *messagesWebservice) readAndEncodePart(part *multipart.Part, maxlen int, destination *string) error {
	var buf []byte
	err := ws.readBlob(part, maxlen, &buf)
	if err != nil {
		return err
	}
	*destination = base64.StdEncoding.EncodeToString(buf)
	return nil
}

func (ws *messagesWebservice) readBlob(part *multipart.Part, maxlen int, destination *[]byte) error {
	var buf bytes.Buffer
	buf.Grow(2 * dao.MEBIBYTE)
	maxlenI64 := int64(maxlen)
	copied, err := io.CopyN(&buf, part, maxlenI64+1)
	if copied > maxlenI64 {
		return fmt.Errorf("part too long: %s", part.FormName())
	}
	if err != nil && err != io.EOF {
		return err
	}
	*destination = buf.Bytes()
	return nil
}

func (ws *messagesWebservice) readEntryFromMultipartBody(request *restful.Request, response *restful.Response) (*dao.MessagesEntry, bool) {
	entry := &dao.MessagesEntry{}

	_, ctParams, err := mime.ParseMediaType(request.HeaderParameter("Content-Type"))
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "couldn't parse content-type header")
		return nil, false
	}
	mr := multipart.NewReader(request.Request.Body, ctParams["boundary"])
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeClientError(response, http.StatusBadRequest, "couldn't parse multipart message")
			return nil, false
		}
		defer part.Close()

		switch part.FormName() {
		case "keySafe":
			err = ws.readAndEncodePart(part, dao.MESSAGE_KEY_SAFE_MAX_BYTES, &entry.KeySafe)
		case "content":
			err = ws.readAndEncodePart(part, dao.MESSAGE_CONTENT_MAX_BYTES, &entry.Content)
		case "meta":
			err = ws.readAndEncodePart(part, dao.MESSAGE_META_MAX_BYTES, &entry.Meta)
		case "attachments":
			err = ws.readBlob(part, dao.MESSAGE_ATTACHMENTS_MAX_BYTES, &entry.Attachments)
		default:
			err = fmt.Errorf("invalid part name: %s", part.FormName())
		}
		if err != nil {
			writeClientError(response, http.StatusBadRequest, err.Error())
			return nil, false
		}
	}

	return entry, true
}

func (ws *messagesWebservice) listEntries(request *restful.Request, response *restful.Response) {
	address, modifiedAfter, includeData, ok := getListParameters(request, response)
	if !ok {
		return
	}

	result := &dataWithResultCount{}
	var rows *sql.Rows
	var err error
	result.ResultsTotal, result.ResultsReturned, rows, err = ws.dao.GetList(address, modifiedAfter, includeData)
	if err != nil {
		writeServerError(err, response)
		return
	}
	defer rows.Close()

	result.Data = make([]interface{}, 0, 100)
	for rows.Next() {
		var entry interface{}
		if includeData {
			entry, err = ws.dao.GetNextEntry(rows)
		} else {
			entry, err = dao.GetNextIDLastModified(rows)
		}
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

	response.WriteEntity(&result)
}

func (ws *messagesWebservice) createEntry(entry *dao.MessagesEntry, request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	authenticated := request.Attribute(AttributeAuthOk)

	entry.Received = time.Now().UTC().Format(time.RFC3339)

	// unauthenticated users are not allowed to set meta
	if authenticated == false {
		entry.Meta = ""
	}

	if !entry.ValidForCreation() {
		writeClientError(response, http.StatusBadRequest, "invalid body sizes")
		return
	}

	err := ws.dao.InsertEntry(address, entry)
	if err != nil {
		writeServerError(err, response)
		return
	}

	if authenticated == true {
		result := &createMessageResult{
			ID:           entry.ID,
			LastModified: entry.LastModified,
			Received:     entry.Received,
		}
		response.WriteEntity(result)
	} else {
		writeEmptyJson(response, http.StatusOK)
	}

	// send push notifications
	notification := notifications.PushNotification{
		Address:        address,
		MessageId:      -1,
		UnreadMessages: -1,
	}
	if authenticated == true {
		// authenticated sending means putting the message in the sender's inbox
		notification.Type = notifications.PushTypeOther
	} else {
		// unauthenticated sending means putting the message in the recipient's inbox
		notification.Type = notifications.PushTypeIncomingMessage
		notification.MessageId = int(entry.ID)
		notification.UnreadMessages = int(ws.dao.GetUnreadCount(address))
	}
	notifications.SendPushNotifications(notification)

	// send email notification(s) if applicable
	if authenticated == false {
		nDao := dao.Notifications{}
		n, err := nDao.GetConfirmedEntry(address)
		switch err {
		case nil:
			language := preferredLanguageFromDb(address)
			notifications.SendMessageNotification(
				address, n.Email, n.WebloginUsername, n.CancelSecret, language)
		case sql.ErrNoRows:
			// no (confirmed) email address found, do nothing
		default:
			util.LogServerError(err)
			return
		}
	}
}

func (ws *messagesWebservice) createEntryFromJson(request *restful.Request, response *restful.Response) {
	entry, ok := ws.readEntryFromJsonBody(request, response)
	if !ok {
		return
	}

	ws.createEntry(entry, request, response)
}

func (ws *messagesWebservice) createEntryFromMultipart(request *restful.Request, response *restful.Response) {
	entry, ok := ws.readEntryFromMultipartBody(request, response)
	if !ok {
		return
	}

	ws.createEntry(entry, request, response)
}

func (ws *messagesWebservice) getEntry(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	id, ok := getID(request, response)
	if !ok {
		return
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

	response.WriteEntity(entry)
}

func (ws *messagesWebservice) modifyMeta(request *restful.Request, response *restful.Response) {
	address, id, lastModified, ok := getModificationParameters(request, response)
	if !ok {
		return
	}

	entry, ok := ws.readEntryFromJsonBody(request, response)
	if !ok {
		return
	}
	entry.SetIDLastModifiedDeleted(id, lastModified, false)

	if !entry.ValidForModification() {
		writeClientError(response, http.StatusBadRequest, "invalid body sizes")
		return
	}

	meta, err := ws.dao.ModifyMeta(address, entry)
	writeEntityOrModificationErr(meta, err, response)
}

func (ws *messagesWebservice) deleteEntry(request *restful.Request, response *restful.Response) {
	address, id, lastModified, ok := getModificationParameters(request, response)
	if !ok {
		return
	}

	meta, err := ws.dao.DeleteEntry(address, id, lastModified)
	writeEntityOrModificationErr(meta, err, response)
}

func (ws *messagesWebservice) getAttachments(request *restful.Request, response *restful.Response) {
	address := request.PathParameter("address")
	id, ok := getID(request, response)
	if !ok {
		return
	}

	attachments, err := ws.dao.GetAttachments(address, id)
	switch {
	case err == sql.ErrNoRows:
		writeClientError(response, http.StatusNotFound, "entry not found")
		return
	case err != nil:
		writeServerError(err, response)
		return
	}

	response.Header().Set(restful.HEADER_ContentType, "application/octet-stream")
	response.Header().Set("Content-Length", strconv.Itoa(len(attachments)))
	response.Write(attachments) //TODO check return values everywhere
}
