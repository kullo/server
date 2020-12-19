/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import "github.com/emicklei/go-restful"

func LanguageFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	acceptLanguage := req.Request.Header.Get("Accept-Language")
	if len(acceptLanguage) > 0 {
		language := chooseLanguage(acceptLanguage)
		req.SetAttribute(AttributeLanguage, language)
	}
	chain.ProcessFilter(req, resp)
}
