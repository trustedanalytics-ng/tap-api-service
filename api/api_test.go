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
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	templateRepositoryModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestValidateImageType(t *testing.T) {
	tests := []struct {
		input   catalogModels.ImageType
		isError bool
	}{
		{"JAVA", false},
		{"GO", false},
		{"NODEJS", false},
		{"PYTHON2.7", false},
		{"PYTHON3.4", false},
		{"", true},
		{"PYTHON", true},
		{"CPP", true},
		{"C", true},
		{"LUA", true},
		{"RUBY", true},
		{"HTML", true},
	}

	Convey("Test validateImageType", t, func() {
		for _, test := range tests {
			Convey(fmt.Sprintf("Given image type %q it should return error response: %v", test.input, test.isError), func() {
				result := validateImageType(test.input)
				if test.isError {
					So(result, ShouldNotBeNil)
				} else {
					So(result, ShouldBeNil)
				}
			})
		}
	})
}

func TestValidateInstanceNumber(t *testing.T) {
	tests := []struct {
		input   int
		isError bool
	}{
		{-1, true},
		{0, false},
		{5, false},
		{6, true},
		{math.MaxInt32, true},
	}

	Convey("Test validateInstanceNumber", t, func() {
		for _, test := range tests {
			Convey(fmt.Sprintf("Given instance number %d it should return error response: %v", test.input, test.isError), func() {
				result := validateInstancesNumber(test.input)
				if test.isError {
					So(result, ShouldNotBeNil)
				} else {
					So(result, ShouldBeNil)
				}
			})
		}
	})
}

func getTestCatalogInstances() []catalogModels.Instance {
	return []catalogModels.Instance{
		{Id: instanceID1, Name: instanceName1, ClassId: serviceID1, Type: catalogModels.InstanceTypeService, Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID1}}},
		{Id: instanceID2, Name: instanceName2, ClassId: serviceID1, Type: catalogModels.InstanceTypeService, Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID1}}},
		{Id: instanceID3, Name: instanceName1, ClassId: serviceID2, Type: catalogModels.InstanceTypeService, Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID1}}},
		{Id: instanceID4, Name: instanceName3, ClassId: serviceID2, Type: catalogModels.InstanceTypeService, Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID1}}},
		{Id: instanceID5, Name: instanceName4, ClassId: serviceID3, Type: catalogModels.InstanceTypeService, Metadata: []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID2}}},
	}
}

func getTestCatalogServices() []catalogModels.Service {
	return []catalogModels.Service{
		{Id: serviceID1, Name: serviceName1, Plans: []catalogModels.ServicePlan{{Id: planID1, Name: planName1}}},
		{Id: serviceID2, Name: serviceName2, Plans: []catalogModels.ServicePlan{{Id: planID1, Name: planName1}}},
		{Id: serviceID3, Name: serviceName3, Plans: []catalogModels.ServicePlan{{Id: planID2, Name: planName2}}},
	}
}

func getTestServiceInstance(instanceID, instanceName, serviceName, serviceID, planName, planID string) models.ServiceInstance {
	return models.ServiceInstance{Id: instanceID, Name: instanceName, OfferingId: serviceID,
		Metadata:    []catalogModels.Metadata{{Id: catalogModels.OFFERING_PLAN_ID, Value: planID}},
		ServiceName: serviceName, ServicePlanName: planName, Type: catalogModels.InstanceTypeService}
}

func TestGetServiceInstances(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)

	testCases := []struct {
		name         string
		offeringName string
		offeringID   string
		planName     string
		result       []models.ServiceInstance
	}{
		{
			name: "", offeringName: "", offeringID: "", planName: "",
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID1, instanceName1, serviceName1, serviceID1, planName1, planID1),
				getTestServiceInstance(instanceID2, instanceName2, serviceName1, serviceID1, planName1, planID1),
				getTestServiceInstance(instanceID3, instanceName1, serviceName2, serviceID2, planName1, planID1),
				getTestServiceInstance(instanceID4, instanceName3, serviceName2, serviceID2, planName1, planID1),
				getTestServiceInstance(instanceID5, instanceName4, serviceName3, serviceID3, planName2, planID2),
			},
		},
		{
			name: strings.ToUpper(instanceName1), offeringName: "", offeringID: "", planName: "",
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID1, instanceName1, serviceName1, serviceID1, planName1, planID1),
				getTestServiceInstance(instanceID3, instanceName1, serviceName2, serviceID2, planName1, planID1),
			},
		},
		{
			name: "", offeringName: strings.ToLower(serviceName1), offeringID: "", planName: "",
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID1, instanceName1, serviceName1, serviceID1, planName1, planID1),
				getTestServiceInstance(instanceID2, instanceName2, serviceName1, serviceID1, planName1, planID1),
			},
		},
		{
			name: "", offeringName: "", offeringID: serviceID2, planName: "",
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID3, instanceName1, serviceName2, serviceID2, planName1, planID1),
				getTestServiceInstance(instanceID4, instanceName3, serviceName2, serviceID2, planName1, planID1),
			},
		},
		{
			name: "", offeringName: "", offeringID: "", planName: strings.ToUpper(planName2),
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID5, instanceName4, serviceName3, serviceID3, planName2, planID2),
			},
		},
		{
			name: instanceName1, offeringName: serviceName2, offeringID: "", planName: "",
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID3, instanceName1, serviceName2, serviceID2, planName1, planID1),
			},
		},
		{
			name: instanceName2, offeringName: serviceName1, offeringID: strings.ToLower(serviceID1), planName: planName1,
			result: []models.ServiceInstance{
				getTestServiceInstance(instanceID2, instanceName2, serviceName1, serviceID1, planName1, planID1),
			},
		},
		{name: wrongInstanceName, offeringName: "", offeringID: "", planName: "", result: []models.ServiceInstance{}},
		{name: "", offeringName: wrongServiceName, offeringID: "", planName: "", result: []models.ServiceInstance{}},
		{name: "", offeringName: "", offeringID: wrongServiceID, planName: "", result: []models.ServiceInstance{}},
		{name: "", offeringName: "", offeringID: "", planName: wrongPlanName, result: []models.ServiceInstance{}},
	}

	Convey(fmt.Sprintf("For sample catalog instances = %v, services = %v and set of filters, GetServiceInstances should return proper responses", getTestCatalogInstances(), getTestCatalogServices()), t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().ListServicesInstances().Return(getTestCatalogInstances(), http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetServices().Return(getTestCatalogServices(), http.StatusOK, nil)

		for _, tc := range testCases {
			Convey(fmt.Sprintf("For filters: name = %q offeringName = %q planName = %q offeringID = %q /api/%s/services should return proper response", tc.name, tc.offeringName, tc.planName, tc.offeringID, apiPrefix), func() {
				url := fmt.Sprintf("/api/%s/%s", apiPrefix, "services")
				// query params should be case insensitive
				url = addToQuery(url, "NAME", tc.name)
				url = addToQuery(url, "ofFeringName", tc.offeringName)
				url = addToQuery(url, "offerIngid", tc.offeringID)
				url = addToQuery(url, "PlanName", tc.planName)

				rr := SendGet(url, mocksAndRouter.router)

				Convey("Status code should be proper", func() {
					if rr.Code != http.StatusOK {
						fmt.Printf("\nbody: %s\n", rr.Body)
					}
					So(rr.Code, ShouldEqual, http.StatusOK)

					Convey("Response should be properly marshaled", func() {
						response := &[]models.ServiceInstance{}
						readAndAssertJson(rr, &response)

						Convey("List of service instances should be proper", func() {
							So(*response, ShouldResemble, tc.result)
						})
					})
				})
			})
		}
	})
}

func TestCreateOffering(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)

	createOfferingMockStructures := arrangeForTestCreateOffering()

	Convey("Given  all structs handled by CreateOffering", t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().AddTemplate(
			createOfferingMockStructures.newCatalogTemplate).Return(
			createOfferingMockStructures.catalogTemplateWithId, http.StatusOK, nil)
		mocksAndRouter.templateRepositoryApiMock.EXPECT().CreateTemplate(
			createOfferingMockStructures.newTemplate).Return(
			http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().UpdateTemplate(
			serviceTemplateID1, []catalogModels.Patch{createOfferingMockStructures.patchUpdateTemplateState}).Return(
			createOfferingMockStructures.catalogTemplateInStateReady, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().AddService(
			createOfferingMockStructures.newCatalogService).Return(
			createOfferingMockStructures.catalogServiceWithId, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().UpdateService(
			serviceID1, []catalogModels.Patch{createOfferingMockStructures.patchUpdateCatalogServiceState}).Return(
			createOfferingMockStructures.catalogServiceInStateReady, http.StatusOK, nil)

		rawTemplate := templateRepositoryModels.RawTemplate{templateRepositoryModels.RAW_TEMPLATE_ID_FIELD: "testId"}
		serviceDeploy := models.ServiceDeploy{Template: rawTemplate, Services: []catalogModels.Service{{
			Name: serviceName1,
			Plans: []catalogModels.ServicePlan{{
				Name: planName1,
			}},
		}}}

		s, err := json.Marshal(serviceDeploy)
		checkTestingError(err)
		buf := bytes.NewBuffer(s)

		r := SendForm(fmt.Sprintf("/api/%s/%s", apiPrefix, "offerings"), buf, "application/json", mocksAndRouter.router)

		Convey("/api/v3/offerings should return proper status code and response", func() {
			So(r.Code, ShouldEqual, http.StatusAccepted)

			response := []catalogModels.Service{}
			readAndAssertJson(r, &response)
			So(response, ShouldResemble, []catalogModels.Service{createOfferingMockStructures.catalogServiceInStateReady})
		})
	})
}

func arrangeForTestCreateOffering() (result createOfferingMockStructures) {
	result = createOfferingMockStructures{}
	result.newCatalogTemplate = catalogModels.Template{State: catalogModels.TemplateStateInProgress}

	result.catalogTemplateWithId = result.newCatalogTemplate
	result.catalogTemplateWithId.Id = serviceTemplateID1

	result.newTemplate = make(templateRepositoryModels.RawTemplate)
	result.newTemplate["id"] = serviceTemplateID1

	patchUpdateTemplateState, err := builder.MakePatchWithPreviousValue("State", catalogModels.TemplateStateReady, catalogModels.TemplateStateInProgress, catalogModels.OperationUpdate)
	checkTestingError(err)
	result.patchUpdateTemplateState = patchUpdateTemplateState

	result.catalogTemplateInStateReady = result.catalogTemplateWithId
	result.catalogTemplateInStateReady.State = catalogModels.TemplateStateReady

	result.newCatalogService = catalogModels.Service{Name: serviceName1, TemplateId: serviceTemplateID1, State: catalogModels.ServiceStateDeploying, Plans: []catalogModels.ServicePlan{{Name: planName1}}}

	result.catalogServiceWithId = result.newCatalogService
	result.catalogServiceWithId.Id = serviceID1

	patchUpdateCatalogServiceState, err := builder.MakePatchWithPreviousValue("State", catalogModels.ServiceStateReady, catalogModels.ServiceStateDeploying, catalogModels.OperationUpdate)
	checkTestingError(err)
	result.patchUpdateCatalogServiceState = patchUpdateCatalogServiceState

	result.catalogServiceInStateReady = result.catalogServiceWithId
	result.catalogServiceInStateReady.State = catalogModels.ServiceStateReady

	return
}
