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
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

func TestBindingToDependence(t *testing.T) {
	Convey("TestBindingToDependence", t, func() {
		mocksAndRouter := prepareMocksAndRouter(t)

		testCatalogServiceInstance := catalogModels.Instance{
			Id:       instanceID1,
			Name:     instanceName1,
			Type:     catalogModels.InstanceTypeService,
			ClassId:  serviceID1,
			Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID1}},
		}
		bindingService := catalogModels.InstanceBindings{Id: instanceID1}

		testCatalogServicePlan := catalogModels.ServicePlan{Id: planID1, Name: planName1}

		testCatalogApplicationInstance := catalogModels.Instance{
			Id:      instanceID2,
			Name:    instanceName2,
			Type:    catalogModels.InstanceTypeApplication,
			ClassId: applicationID1,
		}

		bindingApplication := catalogModels.InstanceBindings{Id: instanceID2}

		expectedDependence := catalogModels.ServiceDependency{
			ServiceId:   serviceID1,
			ServiceName: serviceName1,
			PlanId:      planID1,
			PlanName:    planName1,
		}
		Convey("For service binding", func() {
			mocksAndRouter.catalogApiMock.EXPECT().GetInstance(instanceID1).Return(testCatalogServiceInstance, http.StatusOK, nil)
			mocksAndRouter.catalogApiMock.EXPECT().GetService(serviceID1).Return(getTestCatalogServices()[0], http.StatusOK, nil)
			mocksAndRouter.catalogApiMock.EXPECT().GetServicePlan(serviceID1, planID1).Return(testCatalogServicePlan, http.StatusOK, nil)

			dependence, status, err := bindingToDependence(bindingService)
			Convey("error should be nil, http status should be 200 ok and dependence should be proper", func() {
				So(err, ShouldBeNil)
				So(status, ShouldEqual, http.StatusOK)
				So(dependence, ShouldResemble, expectedDependence)
			})
		})

		Convey("For application binding", func() {
			mocksAndRouter.catalogApiMock.EXPECT().GetInstance(instanceID2).Return(testCatalogApplicationInstance, http.StatusOK, nil)

			dependence, status, err := bindingToDependence(bindingApplication)
			Convey("error should not be nil, http status should be 400 bad request and dependence should be empty", func() {
				So(err, ShouldNotBeNil)
				So(status, ShouldEqual, http.StatusBadRequest)
				So(dependence, ShouldResemble, catalogModels.ServiceDependency{})
			})
		})

		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}
