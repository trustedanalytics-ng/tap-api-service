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
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/twinj/uuid"

	"github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func TestInviteUser(t *testing.T) {
	Convey("User", t, func() {
		mocksAndRouter := prepareMocksAndRouter(t)

		Convey("should be invited when correct email address is applied", func() {
			mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)
			mocksAndRouter.userManagementApiMock.EXPECT().InviteUser(gomock.Any()).Return(nil, 0, nil)
			Convey("Prepare invitation request", func() {
				reqBody, err := json.Marshal(user_management_connector.InvitationRequest{Email: "example@mailinator.com"})
				So(err, ShouldBeNil)

				Convey("Invite user", func() {
					response := commonHttp.SendRequest(http.MethodPost, "/api/v3/users/invitations", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusCreated)
				})
			})
		})
		Convey("should not be invited for empty email address", func() {
			Convey("Prepare invitation request", func() {
				reqBody, err := json.Marshal(user_management_connector.InvitationRequest{Email: ""})
				So(err, ShouldBeNil)
				Convey("Return error", func() {
					response := commonHttp.SendRequest(http.MethodPost, "/api/v3/users/invitations", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Convey("should be invited only when email address is applied", func() {
			Convey("Prepare invitation request", func() {
				reqBody, err := json.Marshal(user_management_connector.InvitationRequest{})
				So(err, ShouldBeNil)
				Convey("Require an email address", func() {
					response := commonHttp.SendRequest(http.MethodPost, "/api/v3/users/invitations", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})
		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestResendUserInvitation(t *testing.T) {
	Convey("User", t, func() {
		mocksAndRouter := prepareMocksAndRouter(t)

		Convey("should be reinvited when correct email address is applied", func() {
			mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)
			mocksAndRouter.userManagementApiMock.EXPECT().ResendUserInvitation(gomock.Any()).Return(0, nil)

			Convey("Prepare invitation request", func() {
				reqBody, err := json.Marshal(user_management_connector.InvitationRequest{Email: "example2@mailinator.com"})
				So(err, ShouldBeNil)

				Convey("Reinvite user", func() {
					response := commonHttp.SendRequest(http.MethodPost, "/api/v3/users/invitations/resend", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusCreated)
				})
			})
		})
		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestListInvitations(t *testing.T) {
	testCases := []struct {
		testInvitations      []string
		getInvitationStatus  int
		expectedErrorMessage string
		errorCase            bool
	}{
		{
			testInvitations:     []string{"invitation1", "invitation2"},
			getInvitationStatus: http.StatusOK,
		},
		{
			errorCase:            true,
			getInvitationStatus:  http.StatusForbidden,
			expectedErrorMessage: "Access Forbidden",
		},
		{
			errorCase:            true,
			getInvitationStatus:  http.StatusInternalServerError,
			expectedErrorMessage: "error doing: Get Invitations",
		},
	}

	Convey("TestListInvitations", t, func() {
		mocksAndRouter, client := prepareMocksAndRouterWithClient(t)

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given testInvitations: %v and user management returns status %v", testCase.testInvitations, testCase.getInvitationStatus), func() {

				mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)

				if testCase.errorCase {
					mocksAndRouter.userManagementApiMock.EXPECT().GetInvitations().Return([]string{}, testCase.getInvitationStatus, errors.New("fake error"))
				} else {
					mocksAndRouter.userManagementApiMock.EXPECT().GetInvitations().Return(testCase.testInvitations, http.StatusOK, nil)
				}

				response, err := client.GetInvitations()

				if testCase.errorCase {
					Convey("GetInvitations should return proper error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, testCase.expectedErrorMessage)
					})
				} else {
					Convey("GetInvitations should return testInvitaions, no error", func() {
						So(err, ShouldBeNil)
						So(response, ShouldResemble, testCase.testInvitations)
					})
				}
			})
		}

		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestListUsers(t *testing.T) {
	testCases := []struct {
		testUsers            []user_management_connector.UaaUser
		getUsersStatus       int
		expectedErrorMessage string
		errorCase            bool
	}{
		{
			testUsers: []user_management_connector.UaaUser{
				{
					Guid:     uuid.NewV4().String(),
					Username: "testUser@example.com",
				},
				{
					Guid:     uuid.NewV4().String(),
					Username: "secondTestUser@example.com",
				},
			},
			getUsersStatus: http.StatusOK,
		},
		{
			errorCase:            true,
			getUsersStatus:       http.StatusForbidden,
			expectedErrorMessage: "Access Forbidden",
		},
		{
			errorCase:            true,
			getUsersStatus:       http.StatusInternalServerError,
			expectedErrorMessage: "error doing: Get Users",
		},
	}

	Convey("TestListUsers", t, func() {
		mocksAndRouter, client := prepareMocksAndRouterWithClient(t)

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given users: %v and given that GetUsers from user-management returns status: %v", testCase.testUsers, testCase.getUsersStatus), func() {

				mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)

				if testCase.errorCase {
					mocksAndRouter.userManagementApiMock.EXPECT().GetUsers().Return([]user_management_connector.UaaUser{}, testCase.getUsersStatus, errors.New("fake error"))
				} else {
					mocksAndRouter.userManagementApiMock.EXPECT().GetUsers().Return(testCase.testUsers, http.StatusOK, nil)
				}

				response, err := client.GetUsers()

				if testCase.errorCase {
					Convey("ListUsers should return proper error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, testCase.expectedErrorMessage)
					})
				} else {
					Convey("ListUsers should return testUsers, no error", func() {
						So(err, ShouldBeNil)
						So(response, ShouldResemble, testCase.testUsers)
					})
				}
			})
		}

		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestDeleteUserInvitation(t *testing.T) {
	testCases := []struct {
		invitationEmail            string
		deleteUserInvitationStatus int
		expectedErrorMessage       string
		errorCase                  bool
	}{
		{
			invitationEmail:            "user@example.com",
			deleteUserInvitationStatus: http.StatusNoContent,
		},
		{
			errorCase:                  true,
			invitationEmail:            "usual-user@example.com",
			deleteUserInvitationStatus: http.StatusForbidden,
			expectedErrorMessage:       "Access Forbidden",
		},
		{
			errorCase:                  true,
			invitationEmail:            "third@example.com",
			deleteUserInvitationStatus: http.StatusInternalServerError,
			expectedErrorMessage:       "error doing: Delete User Invitation",
		},
	}

	Convey("TestDeleteUserInvitation", t, func() {
		mocksAndRouter, client := prepareMocksAndRouterWithClient(t)

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given invitation Email: %v and given that DeleteUserInvitation from user-management returns status: %v", testCase.invitationEmail, testCase.deleteUserInvitationStatus), func() {

				mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)

				if testCase.errorCase {
					mocksAndRouter.userManagementApiMock.EXPECT().DeleteUserInvitation(testCase.invitationEmail).Return(testCase.deleteUserInvitationStatus, errors.New("fake error"))
				} else {
					mocksAndRouter.userManagementApiMock.EXPECT().DeleteUserInvitation(testCase.invitationEmail).Return(http.StatusNoContent, nil)
				}

				err := client.DeleteInvitation(testCase.invitationEmail)

				if testCase.errorCase {
					Convey("DeleteInvitation should return proper error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, testCase.expectedErrorMessage)
					})
				} else {
					Convey("DeleteInvitation should return no error", func() {
						So(err, ShouldBeNil)
					})
				}
			})
		}

		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestDeleteUser(t *testing.T) {
	testCases := []struct {
		userEmail            string
		deleteUserStatus     int
		expectedErrorMessage string
		errorCase            bool
	}{
		{
			userEmail:        "user@example.com",
			deleteUserStatus: http.StatusNoContent,
		},
		{
			errorCase:            true,
			userEmail:            "usual-user@example.com",
			deleteUserStatus:     http.StatusForbidden,
			expectedErrorMessage: "Access Forbidden",
		},
		{
			errorCase:            true,
			userEmail:            "third@example.com",
			deleteUserStatus:     http.StatusInternalServerError,
			expectedErrorMessage: "error doing: Delete User",
		},
	}

	Convey("TestDeleteUser", t, func() {
		mocksAndRouter, client := prepareMocksAndRouterWithClient(t)

		for _, testCase := range testCases {
			Convey(fmt.Sprintf("Given user's email: %v and given that DeleteUser from user-management returns status: %v", testCase.userEmail, testCase.deleteUserStatus), func() {

				mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)

				if testCase.errorCase {
					mocksAndRouter.userManagementApiMock.EXPECT().DeleteUser(testCase.userEmail).Return(testCase.deleteUserStatus, errors.New("fake error"))
				} else {
					mocksAndRouter.userManagementApiMock.EXPECT().DeleteUser(testCase.userEmail).Return(http.StatusNoContent, nil)
				}

				err := client.DeleteUser(testCase.userEmail)

				if testCase.errorCase {
					Convey("DeleteUser should return proper error", func() {
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldContainSubstring, testCase.expectedErrorMessage)
					})
				} else {
					Convey("DeleteUser should return no error", func() {
						So(err, ShouldBeNil)
					})
				}
			})
		}

		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func TestChangeCurrentUserPassword(t *testing.T) {
	Convey("User", t, func() {
		mocksAndRouter := prepareMocksAndRouter(t)
		Convey("should be albe to change current password", func() {
			mocksAndRouter.userManagementApiFactoryMock.EXPECT().GetConfiguredUserManagementConnector(gomock.Any()).Return(mocksAndRouter.userManagementApiMock)
			mocksAndRouter.userManagementApiMock.EXPECT().ChangeCurrentUserPassword(gomock.Any()).Return(0, nil)

			Convey("Prepare change password request", func() {
				reqBody, err := json.Marshal(user_management_connector.ChangePasswordRequest{CurrentPasswd: "donotchange", NewPasswd: "newdonotchange"})
				So(err, ShouldBeNil)

				Convey("Change password", func() {
					response := commonHttp.SendRequest(http.MethodPut, "/api/v3/users/current/password", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusOK)
				})
			})
		})
		Convey("should not be able to change password without new one", func() {
			Convey("Prepare change password request", func() {
				reqBody, err := json.Marshal(user_management_connector.ChangePasswordRequest{CurrentPasswd: "donotchange"})
				So(err, ShouldBeNil)

				Convey("send request", func() {
					response := commonHttp.SendRequest(http.MethodPut, "/api/v3/users/current/password", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})
		Convey("should not be able to change password without declaring current and new", func() {
			Convey("Prepare change password request", func() {
				reqBody, err := json.Marshal(user_management_connector.ChangePasswordRequest{})
				So(err, ShouldBeNil)

				Convey("send request", func() {
					response := commonHttp.SendRequest(http.MethodPut, "/api/v3/users/current/password", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})

		Convey("should not be able to change password without declaring currnet", func() {
			Convey("Prepare change password request", func() {
				reqBody, err := json.Marshal(user_management_connector.ChangePasswordRequest{NewPasswd: "newdonotchange"})
				So(err, ShouldBeNil)

				Convey("send request", func() {
					response := commonHttp.SendRequest(http.MethodPut, "/api/v3/users/current/password", reqBody, mocksAndRouter.router, t)
					So(response.Code, ShouldEqual, http.StatusBadRequest)
				})
			})
		})
		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}
