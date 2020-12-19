/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/util"
	"github.com/emicklei/go-restful"
	"golang.org/x/text/language"
)

const (
	http500Message = "This is probably our fault. We're sorry :-("
)

type dataWithResultCount struct {
	ResultsTotal    uint32        `json:"resultsTotal"`
	ResultsReturned uint32        `json:"resultsReturned"`
	Data            []interface{} `json:"data"`
}

type errorResponseBody struct {
	Status  string `json:"httpStatus"`
	Message string `json:"error"`
}

var ErrBadRequestBodyFormat = errors.New("error in request body format")

var defaultLanguage language.Tag
var languageMatcher language.Matcher

func writeEmptyJson(response *restful.Response, status int) {
	response.WriteHeaderAndEntity(status, make(map[string]string))
}

func newErrorResponseBody(status int, message string) *errorResponseBody {
	return &errorResponseBody{
		Status:  strconv.Itoa(status) + " " + http.StatusText(status),
		Message: message,
	}
}

func writeClientError(response *restful.Response, status int, message string) {
	log.Print("Client error: " + message)
	response.WriteHeaderAndEntity(status, newErrorResponseBody(status, message))
}

func writeRequestValidationError(response *restful.Response, err error) {
	writeClientError(response, http.StatusBadRequest, err.Error())
}

func writeServerError(err error, response *restful.Response) {
	util.LogServerError(err)
	response.WriteHeaderAndEntity(
		http.StatusInternalServerError,
		newErrorResponseBody(http.StatusInternalServerError, http500Message))
}

func boolFromString(str string) (bool, error) {
	if str == "" {
		return false, nil
	}
	return strconv.ParseBool(str)
}

func getModifiedAfter(request *restful.Request, response *restful.Response) (uint64, bool) {
	modifiedAfterStr := request.QueryParameter("modifiedAfter")
	if modifiedAfterStr == "" {
		return 0, true
	}
	modifiedAfter, err := strconv.ParseUint(modifiedAfterStr, 10, 64)
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "bad value for modifiedAfter")
		return 0, false
	}
	return modifiedAfter, true
}

func getLastModified(request *restful.Request, response *restful.Response) (uint64, bool) {
	lastModifiedStr := request.QueryParameter("lastModified")
	lastModified, err := strconv.ParseUint(lastModifiedStr, 10, 64)
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "bad value for lastModified")
		return 0, false
	}
	return lastModified, true
}

func getID(request *restful.Request, response *restful.Response) (uint32, bool) {
	id64, err := strconv.ParseInt(request.PathParameter("id"), 10, 32)
	id := uint32(id64)
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "bad value for id")
		return 0, false
	}
	return id, true
}

func getListParameters(request *restful.Request, response *restful.Response) (string, uint64, bool, bool) {
	address := request.PathParameter("address")
	modifiedAfter, ok := getModifiedAfter(request, response)
	if !ok {
		return "", 0, false, false
	}
	includeData, err := boolFromString(request.QueryParameter("includeData"))
	if err != nil {
		writeClientError(response, http.StatusBadRequest, "bad value for includeData")
		return "", 0, false, false
	}
	return address, modifiedAfter, includeData, true
}

func getModificationParameters(request *restful.Request, response *restful.Response) (string, uint32, uint64, bool) {
	address := request.PathParameter("address")
	id, idOk := getID(request, response)
	lastModified, lastModifiedOk := getLastModified(request, response)
	if !idOk || !lastModifiedOk {
		return "", 0, 0, false
	}
	return address, id, lastModified, true
}

func writeEntityOrModificationErr(body interface{}, err error, response *restful.Response) {
	switch {
	case err == dao.ErrConflict:
		response.WriteHeaderAndEntity(http.StatusConflict, body)
	case err == sql.ErrNoRows:
		writeClientError(response, http.StatusNotFound, "entry not found")
	case err != nil:
		writeServerError(err, response)
	default:
		response.WriteEntity(body)
	}
}

func SetAvailableLanguages(languages ...language.Tag) {
	defaultLanguage = languages[0]
	languageMatcher = language.NewMatcher(languages)
}

func chooseLanguage(acceptLanguage string) string {
	if languageMatcher == nil {
		panic("Must call SetAvailableLanguages in initialization")
	}
	tags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
	chosenTag, _, _ := languageMatcher.Match(tags...)
	return chosenTag.String()
}

func preferredLanguage(req *restful.Request) string {
	language, ok := req.Attribute(AttributeLanguage).(string)
	if ok {
		return language
	} else {
		return defaultLanguage.String()
	}
}

func preferredLanguageFromDb(address string) string {
	users := dao.Users{}
	recipient, err := users.GetEntry(address)
	if err != nil {
		util.LogServerError(err)
		return defaultLanguage.String()
	} else if recipient.Language != "" {
		return recipient.Language
	} else {
		return defaultLanguage.String()
	}
}
