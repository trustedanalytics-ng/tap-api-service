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
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gocraft/web"

	uaaConnector "github.com/trustedanalytics-ng/tap-api-service/uaa-connector"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const (
	adminGroup = "tap.admin"
	userGroup  = "tap.user"
)

func (c *Context) BasicAuthorizeMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	logger.Info("Trying to access url ", req.URL.Path, " by BasicAuthorize")
	username, password, is_ok := req.BasicAuth()

	if !is_ok || username != os.Getenv("API_SERVICE_USER") || password != os.Getenv("API_SERVICE_PASS") {
		logger.Info("Invalid Basic Auth credentials")
		commonHttp.RespondUnauthorized(rw)
		return
	}
	logger.Info("User authenticated as ", username)
	next(rw, req)
}

func (c *Context) Oauth2AuthorizeMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.authorizeUser([]string{userGroup, adminGroup}, rw, req, next)
}

func (c *Context) Oauth2AuthorizeAdminMiddleware(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	c.authorizeUser([]string{adminGroup}, rw, req, next)
}

func (c *Context) authorizeUser(allowedRoles []string, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	jwt, err := getJwtToken(req)
	if err != nil {
		commonHttp.RespondUnauthorized(rw)
		return
	}

	if !hasRole(jwt.Scope, allowedRoles) {
		logger.Infof("Endpoint only for admins, %v is not admin", jwt.Username)
		commonHttp.Respond403(rw)
		return
	}

	c.Username = jwt.Username
	next(rw, req)
}

func getJwtToken(req *web.Request) (*uaaConnector.TapJWTToken, error) {
	logger.Info("Trying to access url ", req.URL.Path, " by OAuth2Authorize")

	authorization := req.Header.Get("Authorization")

	if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		err := errors.New("no bearer in Authorization header")
		return nil, err
	}

	token := authorization[7:]

	jwt, err := BrokerConfig.UaaApi.ValidateOauth2Token(token)
	if err != nil {
		err := errors.New("Invalid auth token")
		return nil, err
	}

	logger.Infof("User authenticated as %v <%v>", jwt.Username, jwt.Email)
	return jwt, nil
}

func hasRole(userRoles []string, allowedRoles []string) bool {
	for _, role := range allowedRoles {
		if commonHttp.StringInSlice(role, userRoles) {
			return true
		}
	}
	return false
}

func (c *Context) Login(rw web.ResponseWriter, req *web.Request) {
	username, password, is_ok := req.BasicAuth()
	if !is_ok {
		commonHttp.Respond400(rw, errors.New("No credentials provided"))
		return
	}

	logger.Infof("Getting token for %s", username)
	loginResp, status, err := BrokerConfig.UaaApi.Login(username, password)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	commonHttp.WriteJson(rw, loginResp, http.StatusOK)
}

func (c *Context) getAuditTrail() catalogModels.AuditTrail {
	return catalogModels.AuditTrail{
		LastUpdateBy: c.Username,
		CreatedBy:    c.Username,
	}
}
