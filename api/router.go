/**
 * Copyright (c) 2016 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package api

import (
	"os"

	"fmt"
	"github.com/gocraft/web"
)

func SetupRouter(r *web.Router, oauthMiddlewareActivated bool) {
	context := getPlatformSettings()
	r.Middleware(web.LoggerMiddleware)
	r.Middleware((context).CheckBrokerConfig)

	r.Get("/api", context.Introduce)
	r.Get("/healthz", context.GetApiServiceHealth)

	context.buildRouteAliasses(r, []string{"v1", "v1.0", "v3", "v3.0"}, oauthMiddlewareActivated)
}

func (c Context) buildRouteAliasses(r *web.Router, aliasses []string, oauthMiddlewareActivated bool) {
	for _, alias := range aliasses {
		aliasString := fmt.Sprintf("/api/%s", alias)
		r.Get(fmt.Sprintf("%s/login", aliasString), c.Login)

		aliasRouter := r.Subrouter(c, aliasString)
		route(aliasRouter, &c, oauthMiddlewareActivated)
	}
}

func getPlatformSettings() Context {
	context := Context{}
	context.Domain = os.Getenv("DOMAIN")
	context.TapVersion = os.Getenv("TAP_VERSION")
	context.CliVersion = os.Getenv("CLI_VERSION")
	context.CdhVersion = os.Getenv("CDH_VERSION")
	context.K8sVersion = os.Getenv("K8S_VERSION")
	context.CoreOrganization = os.Getenv("CORE_ORGANIZATION")
	return context
}
