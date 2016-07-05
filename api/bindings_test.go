/**
 * Copyright (c) 2017 Intel Corporation
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
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/twinj/uuid"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

type testReturn struct {
	ID     string
	Status int
	Error  error
}

func TestGetInstanceIdBasedOnOneOfIds(t *testing.T) {
	ctx := Context{}
	testingPack := prepareMocksAndRouter(t)

	properAppInstanceID := uuid.NewV4().String()
	properAppInstanceID2 := uuid.NewV4().String()
	properAppID := uuid.NewV4().String()
	properSvcInstanceID := uuid.NewV4().String()
	emptyID := ""

	Convey("TestGetInstanceIdBasedOnOneOfIds", t, func() {
		Convey("Given request with fulfilled Application ID", func() {
			cases := map[string]struct {
				instanceList []catalogModels.Instance
				ret          testReturn
			}{
				"when instance in Catalog, should return proper Instance ID": {
					[]catalogModels.Instance{{Id: properAppInstanceID}},
					testReturn{
						ID:     properAppInstanceID,
						Status: http.StatusOK,
						Error:  nil},
				},

				"when instance not in Catalog, should throw an error": {
					[]catalogModels.Instance{},
					testReturn{
						ID:     emptyID,
						Status: http.StatusNotFound,
						Error:  fmt.Errorf(errNoInstanceOfApplication, properAppID)},
				},

				"when two instances in Catalog, should throw an error": {
					[]catalogModels.Instance{{Id: properAppInstanceID}, {Id: properAppInstanceID2}},
					testReturn{
						ID:     emptyID,
						Status: http.StatusInternalServerError,
						Error:  fmt.Errorf(errTooManyInstanceOfApplication, properAppID)},
				},
			}

			for description, test := range cases {
				Convey(description, func() {
					testingPack.catalogApiMock.EXPECT().
						ListApplicationInstances(properAppID).
						Return(test.instanceList, http.StatusOK, nil)

					InstanceID, status, err := ctx.getInstanceIdBasedOnOneOfIds(&models.InstanceBindingRequest{
						ApplicationId: properAppID,
					})

					So(err, ShouldResemble, test.ret.Error)
					So(status, ShouldEqual, test.ret.Status)

					if test.ret.ID != "" {
						So(InstanceID, ShouldEqual, test.ret.ID)
					}

				})
			}
		})

		GetInstanceMockStatus := func(instance catalogModels.Instance, status int, err error) func() {
			return func() {
				testingPack.catalogApiMock.EXPECT().
					GetInstance(properSvcInstanceID).
					Return(instance, status, err)
			}
		}

		GetInstanceMock := func(instance catalogModels.Instance) func() {
			return GetInstanceMockStatus(instance, http.StatusOK, nil)
		}

		Convey("Given request with fulfilled Service ID", func() {
			cases := map[string]struct {
				mock func()
				ret  testReturn
			}{
				"when instance in Catalog, should return proper instance ID": {
					GetInstanceMock(catalogModels.Instance{Id: properSvcInstanceID, Type: catalogModels.InstanceTypeService}),
					testReturn{
						ID:     properSvcInstanceID,
						Status: http.StatusOK,
						Error:  nil},
				},

				"when instance not in Catalog, should throw an error": {
					GetInstanceMockStatus(catalogModels.Instance{}, http.StatusNotFound, errors.New("Not Found")),
					testReturn{
						ID:     emptyID,
						Status: http.StatusNotFound,
						Error:  fmt.Errorf(errNoServiceInstanceInCatalog, properSvcInstanceID, "Not Found")},
				},

				"when instance in Catalog, but is not a Service, should throw an error Error": {
					GetInstanceMock(catalogModels.Instance{Id: properSvcInstanceID, Type: catalogModels.InstanceTypeApplication}),
					testReturn{
						ID:     emptyID,
						Status: http.StatusNotFound,
						Error:  fmt.Errorf(errInstanceIsNotAService, properSvcInstanceID)},
				},
			}

			for description, test := range cases {
				Convey(description, func() {
					test.mock()

					InstanceID, status, err := ctx.getInstanceIdBasedOnOneOfIds(&models.InstanceBindingRequest{
						ServiceId: properSvcInstanceID,
					})

					So(err, ShouldResemble, test.ret.Error)
					So(status, ShouldEqual, test.ret.Status)

					if test.ret.ID != "" {
						So(InstanceID, ShouldEqual, test.ret.ID)
					}

				})
			}
		})

		Convey("Given request with fulfilled Application ID and Service ID", func() {
			Convey("should return error", func() {
				InstanceID, status, err := ctx.getInstanceIdBasedOnOneOfIds(&models.InstanceBindingRequest{
					ServiceId:     properSvcInstanceID,
					ApplicationId: properAppID,
				})

				So(err, ShouldResemble, errBadInstanceBindingRequest)
				So(status, ShouldEqual, http.StatusBadRequest)

				if emptyID != "" {
					So(InstanceID, ShouldEqual, emptyID)
				}

			})
		})

		Convey("Given empty request", func() {
			Convey("should return error", func() {
				InstanceID, status, err := ctx.getInstanceIdBasedOnOneOfIds(&models.InstanceBindingRequest{})

				So(err, ShouldResemble, errBadInstanceBindingRequest)
				So(status, ShouldEqual, http.StatusBadRequest)

				if emptyID != "" {
					So(InstanceID, ShouldEqual, emptyID)
				}

			})
		})
	})
}
