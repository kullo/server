/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"crypto/sha512"
	"encoding/base64"
	"net/http"
	"strings"

	"bitbucket.org/kullo/server/dao"
	"github.com/emicklei/go-restful"
)

const AttributeAuthOk = "authOk"
const AttributeLanguage = "language"

func checkLoginKey(address, loginKey string) (bool, error) {
	loginKeyHash := sha512.Sum512([]byte(loginKey))
	loginKeyBase64 := base64.StdEncoding.EncodeToString(loginKeyHash[:])

	var users dao.Users
	exists, err := users.UserExistsAndActive(address)
	if !exists || err != nil {
		return false, err
	}

	keysSymm := dao.KeysSymm{}
	entry, err := keysSymm.GetEntry(address)
	switch {
	case err != nil:
		return false, err
	}

	return loginKeyBase64 == entry.LoginKey, nil
}

func checkAuthnAndAuthz(authHeader, expectedAddress string) (bool, error) {
	if expectedAddress == "" {
		return false, nil
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return false, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, nil
	}
	addressAndLoginKey := strings.SplitN(string(decoded), ":", 2)
	if len(addressAndLoginKey) != 2 {
		return false, nil
	}
	address := addressAndLoginKey[0]
	loginKey := addressAndLoginKey[1]
	if expectedAddress != address {
		return false, nil
	}
	return checkLoginKey(address, loginKey)
}

func doCheckAuth(req *restful.Request) (bool, error) {
	authHeader := req.Request.Header.Get("Authorization")
	address := req.PathParameter("address")

	if len(authHeader) > 0 {
		authOk, err := checkAuthnAndAuthz(authHeader, address)
		if err == nil && authOk {
			var users = dao.Users{}
			language, ok := req.Attribute(AttributeLanguage).(string)
			if ok {
				err = users.UpdateLatestLoginTimestampAndLanguage(address, language)
			} else {
				err = users.UpdateLatestLoginTimestamp(address)
			}
		}
		return authOk, err
	} else {
		return false, nil
	}
}

func AuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	authOk, err := doCheckAuth(req)
	if err != nil {
		writeServerError(err, resp)
		return
	}
	if !authOk {
		resp.AddHeader("WWW-Authenticate", "Basic realm=Kullo")
		writeClientError(resp, http.StatusUnauthorized, "not authorized")
		return
	}
	chain.ProcessFilter(req, resp)
}

func OptionalAuthFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	authOk, err := doCheckAuth(req)
	if err != nil {
		writeServerError(err, resp)
		return
	}
	req.SetAttribute(AttributeAuthOk, authOk)
	chain.ProcessFilter(req, resp)
}
