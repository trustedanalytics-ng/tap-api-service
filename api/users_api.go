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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gocraft/web"
	"gopkg.in/validator.v2"

	userManagement "github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *Context) InviteUser(rw web.ResponseWriter, req *web.Request) {
	invReq := userManagement.InvitationRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	err = json.Unmarshal(body, &invReq)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	if err := validator.Validate(invReq); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	authorization := req.Header.Get("Authorization")
	resp, status, err := BrokerConfig.
		UserManagementApiFactory.
		GetConfiguredUserManagementConnector(authorization).
		InviteUser(invReq.Email)
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Invite User")
		return
	}

	commonHttp.WriteJson(rw, resp, http.StatusCreated)
}

func (c *Context) ResendUserInvitation(rw web.ResponseWriter, req *web.Request) {
	invReq := userManagement.InvitationRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	err = json.Unmarshal(body, &invReq)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	if err := validator.Validate(invReq); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	authorization := req.Header.Get("Authorization")
	status, err := BrokerConfig.
		UserManagementApiFactory.
		GetConfiguredUserManagementConnector(authorization).
		ResendUserInvitation(invReq.Email)
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Resend User Invitation")
		return
	}

	commonHttp.WriteJson(rw, "", http.StatusCreated)
}

func (c *Context) ListInvitations(rw web.ResponseWriter, req *web.Request) {
	authorization := req.Header.Get("Authorization")
	resp, status, err := BrokerConfig.UserManagementApiFactory.GetConfiguredUserManagementConnector(authorization).GetInvitations()
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Get Invitations")
		return
	}

	commonHttp.WriteJson(rw, resp, http.StatusOK)
}

func (c *Context) ListUsers(rw web.ResponseWriter, req *web.Request) {
	authorization := req.Header.Get("Authorization")
	resp, status, err := BrokerConfig.UserManagementApiFactory.GetConfiguredUserManagementConnector(authorization).GetUsers()
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Get Users")
		return
	}

	commonHttp.WriteJson(rw, resp, http.StatusOK)
}

func (c *Context) DeleteUserInvitation(rw web.ResponseWriter, req *web.Request) {
	invReq := userManagement.InvitationRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	err = json.Unmarshal(body, &invReq)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	authorization := req.Header.Get("Authorization")
	status, err := BrokerConfig.UserManagementApiFactory.GetConfiguredUserManagementConnector(authorization).DeleteUserInvitation(invReq.Email)
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Delete User Invitation")
		return
	}

	commonHttp.WriteJson(rw, "", status)
}

func (c *Context) DeleteUser(rw web.ResponseWriter, req *web.Request) {
	invReq := userManagement.InvitationRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	err = json.Unmarshal(body, &invReq)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	authorization := req.Header.Get("Authorization")
	status, err := BrokerConfig.UserManagementApiFactory.GetConfiguredUserManagementConnector(authorization).DeleteUser(invReq.Email)
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Delete User")
		return
	}

	commonHttp.WriteJson(rw, "", status)
}

func (c *Context) ChangeCurrentUserPassword(rw web.ResponseWriter, req *web.Request) {
	changePasswdReq := userManagement.ChangePasswordRequest{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	err = json.Unmarshal(body, &changePasswdReq)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	if err := validator.Validate(changePasswdReq); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	authorization := req.Header.Get("Authorization")
	status, err := BrokerConfig.
		UserManagementApiFactory.
		GetConfiguredUserManagementConnector(authorization).
		ChangeCurrentUserPassword(changePasswdReq)
	if err != nil {
		commonHttp.RespondErrorByStatus(rw, status, "Change Current User Password")
		return
	}

	commonHttp.WriteJson(rw, "", http.StatusOK)
}
