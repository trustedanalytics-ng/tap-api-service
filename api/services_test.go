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

	apiServiceModels "github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

func TestParseCatalogServiceToApiService(t *testing.T) {
	testPlans := []catalogModels.ServicePlan{{
		Id:          planID1,
		Name:        planName1,
		Description: planDescription1,
		Cost:        planCost1,
	}, {
		Id:          planID2,
		Name:        planName2,
		Description: planDescription2,
		Cost:        planCost2,
	}}
	testMetadata := []catalogModels.Metadata{{
		Id:    metadataID1,
		Value: metadataValue1,
	}, {
		Id:    metadataID2,
		Value: metadataValue2,
	}}

	testCatalogService := catalogModels.Service{
		Id:          serviceID1,
		Name:        serviceName1,
		Description: serviceDescription1,
		Bindable:    serviceBindable1,
		TemplateId:  serviceTemplateID1,
		State:       serviceState1,
		Plans:       testPlans,
		Metadata:    testMetadata,
	}

	expectedApiService := apiServiceModels.Offering{
		Name:        serviceName1,
		Description: serviceDescription1,
		Bindable:    serviceBindable1,
		Id:          serviceID1,
		State:       string(serviceState1),
		Metadata:    testMetadata,
		OfferingPlans: []apiServiceModels.OfferingPlan{
			apiServiceModels.OfferingPlan{
				Name:        planName1,
				Description: planDescription1,
				OfferingId:  serviceID1,
				Id:          planID1,
				Active:      true,
				Free:        true,
			},
			apiServiceModels.OfferingPlan{
				Name:        planName2,
				Description: planDescription2,
				OfferingId:  serviceID1,
				Id:          planID2,
				Active:      true,
				Free:        false,
			},
		},
	}

	Convey("For sample catalog service ParseCatalogServiceToApiService should return proper api service structure", t, func() {
		apiService := ParseServiceToOffering(testCatalogService, nil)

		So(apiService, ShouldResemble, expectedApiService)
	})
}

func TestValidateIfBrokerNameCanBeUsed(t *testing.T) {
	Convey("Test testValidateIfBrokerNameCanBeUsed method", t, func() {
		mocksAndRouter := prepareMocksAndRouter(t)

		Convey("should return error if broker name is empty", func() {
			err := validateIfBrokerNameCanBeUsed("")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "broker name cannont be empty!")

		})

		Convey("should return error if broker name is incorrect", func() {
			err := validateIfBrokerNameCanBeUsed("wrong^-_name")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "broker name is invalid: ")

		})

		Convey("should return error if broker name is in use", func() {
			instances := []catalogModels.Instance{{Name: "used"}}

			mocksAndRouter.catalogApiMock.EXPECT().ListInstances().Return(instances, http.StatusOK, nil)

			err := validateIfBrokerNameCanBeUsed("used")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "brokerName used is already in use!")

		})

		Convey("should return nil if broker name is correct", func() {
			instances := []catalogModels.Instance{{Name: "pl"}}

			mocksAndRouter.catalogApiMock.EXPECT().ListInstances().Return(instances, http.StatusOK, nil)

			err := validateIfBrokerNameCanBeUsed("ok")
			So(err, ShouldBeNil)

		})
		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}
