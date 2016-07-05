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
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics-ng/tap-api-service/uaa-connector"
	"github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	"github.com/trustedanalytics-ng/tap-catalog/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const (
	userAllowedUrl    = "/api/v3/offerings"
	adminAllowedUrl   = "/api/v3/users/invitations"
	testToken         = "TEST_TOKEN"
	fakeEmail         = "email.address@test.com"
	fakeEmailJsonBody = `{"email": "email.address@test.com"}`
)

type testCase struct {
	url            string
	token          uaa_connector.TapJWTToken
	expectedStatus int
	body           string
}

func TestOauth2AuthorizeMiddleware(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouterWithOauth2Activated(t)

	Convey("Given a request for the user endpoint and a user access token", t, func() {
		testCases := []testCase{
			{
				url:            userAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"test.scope", "tap.user"}},
				expectedStatus: http.StatusOK,
			},
			{
				url:            userAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"test.scope", "tap.admin"}},
				expectedStatus: http.StatusOK,
			},
			{
				url:            userAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"tap.user", "tap.admin"}},
				expectedStatus: http.StatusOK,
			},
			{
				url:            userAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"test.scope", "test.scope2"}},
				expectedStatus: http.StatusForbidden,
			},
		}

		// mock /api/v3/offerings
		mocksAndRouter.catalogApiMock.EXPECT().GetServices().Return([]models.Service{}, 0, nil)

		header := http.Header{}
		header.Set("Content-Type", "application/json")
		header.Set("Authorization", fmt.Sprintf("bearer %s", testToken))

		emptyBody := []byte{}

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given a token scope %v TAP should return code %v", testCase.token.Scope, testCase.expectedStatus), func() {
				mocksAndRouter.uaaApiMock.EXPECT().ValidateOauth2Token(testToken).Return(&testCase.token, nil)

				response := commonHttp.SendRequestWithHeaders("GET", testCase.url, emptyBody, mocksAndRouter.router, header, t)
				So(response.Code, ShouldEqual, testCase.expectedStatus)
			})
		}
	})
}

func TestOauth2AuthorizeAdminMiddleware(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouterWithOauth2Activated(t)

	Convey("Given a request for the admin endpoint and a user access token", t, func() {
		testCases := []testCase{
			{
				url:            adminAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"test.scope", "tap.admin"}},
				expectedStatus: http.StatusCreated,
				body:           fakeEmailJsonBody,
			},
			{
				url:            adminAllowedUrl,
				token:          uaa_connector.TapJWTToken{Scope: []string{"test.scope", "tap.user"}},
				expectedStatus: http.StatusForbidden,
				body:           fakeEmailJsonBody,
			},
		}

		invitationResponse := user_management_connector.InvitationResponse{}
		authorization := fmt.Sprintf("bearer %s", testToken)

		// mock POST /api/v3/users/invitations
		mocksAndRouter.userManagementApiMock.EXPECT().InviteUser(fakeEmail).Return(&invitationResponse, 0, nil)
		mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(authorization).Return(mocksAndRouter.userManagementApiMock)

		header := http.Header{}
		header.Set("Content-Type", "application/json")
		header.Set("Authorization", authorization)

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given a token scope %v TAP should return code %v", testCase.token.Scope, testCase.expectedStatus), func() {
				mocksAndRouter.uaaApiMock.EXPECT().ValidateOauth2Token(testToken).Return(&testCase.token, nil)

				response := commonHttp.SendRequestWithHeaders("POST", testCase.url, []byte(testCase.body), mocksAndRouter.router, header, t)
				So(response.Code, ShouldEqual, testCase.expectedStatus)
			})
		}

	})
}
