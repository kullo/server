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

func UserFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	address := req.PathParameter("address")

	var users dao.Users
	exists, err := users.UserExistsAndActive(address)
	if err != nil {
		writeServerError(err, resp)
		return
	}
	if !exists {
		writeClientError(resp, http.StatusNotFound, "user not found")
		return
	}
	chain.ProcessFilter(req, resp)
}
