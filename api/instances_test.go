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
	"fmt"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/twinj/uuid"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	commonUtil "github.com/trustedanalytics-ng/tap-go-common/util"
)

const (
	randomNamesLength       = 15
	randomDescriptionLength = 60
)

type CaseDataSet struct {
	app                           catalogModels.Application
	instance                      catalogModels.Instance
	image                         catalogModels.Image
	expectedApiServiceAppInstance models.ApplicationInstance
}

func createCase(stateOfInstance catalogModels.InstanceState, stateOfImage catalogModels.ImageState) (data CaseDataSet) {
	ID := uuid.NewV4().String()
	app := catalogModels.Application{
		Id:                   ID,
		Name:                 "name-" + ID,
		Description:          "some-desc",
		ImageId:              catalogModels.USER_DEFINED_APPLICATION_IMAGE_PREFIX + ID,
		Replication:          1,
		TemplateId:           uuid.NewV4().String(),
		InstanceDependencies: []catalogModels.InstanceDependency{},
	}

	instance := catalogModels.Instance{
		Id:       ID,
		Name:     app.Name,
		Type:     catalogModels.InstanceTypeApplication,
		ClassId:  app.Id,
		Bindings: []catalogModels.InstanceBindings{},
		State:    stateOfInstance,
	}

	imageOfApp := catalogModels.Image{
		Id:       catalogModels.USER_DEFINED_APPLICATION_IMAGE_PREFIX + app.Id,
		Type:     catalogModels.ImageTypeJava,
		BlobType: catalogModels.BlobTypeJar,
		State:    stateOfImage,
	}

	expectedApiServiceAppInstance := models.ApplicationInstance{
		Id:               instance.Id,
		Name:             app.Name,
		Type:             instance.Type,
		Bindings:         instance.Bindings,
		Metadata:         instance.Metadata,
		State:            getAppInstanceState(instance.State, imageOfApp.State),
		AuditTrail:       instance.AuditTrail,
		Replication:      app.Replication,
		ImageState:       imageOfApp.State,
		ImageType:        imageOfApp.Type,
		Memory:           apiApplicationInstanceMemoryDefault,
		DiskQuota:        apiApplicationInstanceDiskQuotaDefault,
		RunningInstances: app.Replication,
		Urls:             []string{""},
	}

	data = CaseDataSet{
		app:      app,
		instance: instance,
		image:    imageOfApp,
		expectedApiServiceAppInstance: expectedApiServiceAppInstance,
	}
	return
}

func createCaseEmptyImageID() (data CaseDataSet) {
	ID := uuid.NewV4().String()
	app := catalogModels.Application{
		Id:                   ID,
		Name:                 "random-name2",
		Description:          "random-desc2",
		Replication:          1,
		TemplateId:           uuid.NewV4().String(),
		InstanceDependencies: []catalogModels.InstanceDependency{},
	}

	instance := catalogModels.Instance{}

	expectedApiServiceAppInstance := models.ApplicationInstance{
		Id:               app.Id,
		Name:             app.Name,
		Type:             instance.Type,
		Bindings:         instance.Bindings,
		Metadata:         instance.Metadata,
		State:            catalogModels.InstanceStateFailure,
		AuditTrail:       instance.AuditTrail,
		Replication:      app.Replication,
		ImageState:       errorFetchImageState,
		ImageType:        errorFetchImageType,
		Memory:           apiApplicationInstanceMemoryDefault,
		DiskQuota:        apiApplicationInstanceDiskQuotaDefault,
		RunningInstances: app.Replication,
		Urls:             []string{""},
	}

	data = CaseDataSet{
		app: app,
		expectedApiServiceAppInstance: expectedApiServiceAppInstance,
	}
	return
}

func arrangeForTestParseToApiApplicationInstances() (dataSets []CaseDataSet) {
	caseData := createCase(catalogModels.InstanceStateRunning, catalogModels.ImageStateReady)
	dataSets = append(dataSets, caseData)

	caseData = createCase("", catalogModels.ImageStateError)
	dataSets = append(dataSets, caseData)

	caseData = createCaseEmptyImageID()
	dataSets = append(dataSets, caseData)

	return
}

func TestParseToApiApplicationInstances(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)
	casesData := arrangeForTestParseToApiApplicationInstances()
	var apps []catalogModels.Application
	var instances []catalogModels.Instance
	var expectedApplicationInstances []models.ApplicationInstance

	for _, caseData := range casesData {
		apps = append(apps, caseData.app)
		instances = append(instances, caseData.instance)
		expectedApplicationInstances = append(expectedApplicationInstances, caseData.expectedApiServiceAppInstance)
	}

	Convey(fmt.Sprint("Given: proper application, application with image in error state, application without imageID"), t, func() {
		for _, caseData := range casesData {
			app := caseData.app
			if app.ImageId != "" {
				mocksAndRouter.catalogApiMock.EXPECT().GetImage(app.ImageId).Return(caseData.image, http.StatusOK, nil)
			}
		}

		realApplicationInstances, err := ParseToApiApplicationInstances(apps, instances)

		Convey(fmt.Sprint("ParseToApiApplicationInstances should not return error"), func() {
			So(err, ShouldBeNil)
		})

		Convey(fmt.Sprint("result of ParseToApiApplicationInstances should resemble expected response"), func() {
			So(realApplicationInstances, ShouldResemble, expectedApplicationInstances)
		})
	})

	data := createCase("", "")

	getImageFakeError := errors.New("not found")
	expectedError := errors.New("Cannot fetch image " + data.app.ImageId + " from Catalog: " + getImageFakeError.Error())

	Convey(fmt.Sprint("Given application with non-existent imageID"), t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().GetImage(data.app.ImageId).Return(catalogModels.Image{}, http.StatusNotFound,
			getImageFakeError)

		_, err := ParseToApiApplicationInstances([]catalogModels.Application{data.app}, []catalogModels.Instance{})

		Convey(fmt.Sprint("ParseToApiApplicationInstances should return expected error"), func() {
			So(err, ShouldResemble, expectedError)
		})
	})
}

func TestGetAppInstanceState(t *testing.T) {
	testCases := []struct {
		instanceState catalogModels.InstanceState
		imageState    catalogModels.ImageState
		result        catalogModels.InstanceState
	}{
		{instanceState: "", imageState: catalogModels.ImageStateError, result: catalogModels.InstanceStateFailure},
		{instanceState: "", imageState: errorFetchImageState, result: catalogModels.InstanceStateFailure},
		{instanceState: "", imageState: "some_image_state", result: waitingForImageApplicationInstanceState},
		{instanceState: "some_state", imageState: "some_image_state", result: "some_state"},
	}

	Convey("For set of test cases getAppInstanceState should return proper responses", t, func() {
		for _, tc := range testCases {
			Convey(fmt.Sprintf("Case of instance state %q and image state %q", tc.instanceState, tc.imageState), func() {
				result := getAppInstanceState(tc.instanceState, tc.imageState)

				So(result, ShouldEqual, tc.result)
			})
		}
	})
}

func generateProperService() catalogModels.Service {
	return catalogModels.Service{
		Id:          uuid.NewV4().String(),
		Name:        commonUtil.RandomString(randomNamesLength),
		Description: commonUtil.RandomString(randomDescriptionLength),
		Bindable:    true,
		TemplateId:  uuid.NewV4().String(),
		State:       catalogModels.ServiceStateReady,
		Plans: []catalogModels.ServicePlan{
			{
				Id:   uuid.NewV4().String(),
				Name: commonUtil.RandomString(randomNamesLength),
			},
			{
				Id:   uuid.NewV4().String(),
				Name: commonUtil.RandomString(randomNamesLength),
			},
			{
				Id:   uuid.NewV4().String(),
				Name: commonUtil.RandomString(randomNamesLength),
			},
		},
	}

}

func TestParseToApiServiceInstances(t *testing.T) {
	Convey(fmt.Sprintf("Given proper services, instance with and without %v metadata entry", catalogModels.OFFERING_PLAN_ID), t, func() {
		services := []catalogModels.Service{}

		for i := 0; i < 5; i++ {
			services = append(services, generateProperService())
		}

		badInstance := catalogModels.Instance{
			Id:      uuid.NewV4().String(),
			ClassId: services[0].Id,
			Type:    catalogModels.InstanceTypeService,
		}

		goodInstance := catalogModels.Instance{
			Id:      uuid.NewV4().String(),
			ClassId: services[1].Id,
			Type:    catalogModels.InstanceTypeService,
			Metadata: []catalogModels.Metadata{
				{
					Id:    catalogModels.OFFERING_PLAN_ID,
					Value: services[1].Plans[0].Id,
				},
			},
		}

		instances := []catalogModels.Instance{badInstance, goodInstance}

		expectedResult := []models.ServiceInstance{
			{
				Id:              goodInstance.Id,
				Name:            goodInstance.Name,
				Type:            goodInstance.Type,
				ServiceName:     services[1].Name,
				ServicePlanName: services[1].Plans[0].Name,
				Bindings:        goodInstance.Bindings,
				Metadata:        goodInstance.Metadata,
				AuditTrail:      goodInstance.AuditTrail,
				State:           goodInstance.State,
				OfferingId:      services[1].Id,
			},
		}

		Convey(fmt.Sprintf("ParseToApiServiceInstances should return only api service instance with %v metadata entry", catalogModels.OFFERING_PLAN_ID), func() {
			result := ParseToApiServiceInstances(services, instances)

			So(result, ShouldResemble, expectedResult)
		})
	})
}

func TestConvertToApiServiceInstance(t *testing.T) {
	Convey("Given proper service", t, func() {
		service := generateProperService()
		instance := catalogModels.Instance{

			Id:      uuid.NewV4().String(),
			ClassId: service.Id,
			Type:    catalogModels.InstanceTypeService,
		}

		Convey(fmt.Sprintf("and instance without %v metadata entry", catalogModels.OFFERING_PLAN_ID), func() {
			Convey("ConvertToApiServiceInstance should return error", func() {
				_, err := ConvertToApiServiceInstance(service, instance)

				So(err, ShouldNotBeNil)
			})

		})

		Convey(fmt.Sprintf("and proper instance with %v metadata entry", catalogModels.OFFERING_PLAN_ID), func() {
			instance.Metadata = []catalogModels.Metadata{
				{
					Id:    catalogModels.OFFERING_PLAN_ID,
					Value: service.Plans[0].Id,
				},
			}

			expectedResult := models.ServiceInstance{
				Id:              instance.Id,
				Name:            instance.Name,
				Type:            instance.Type,
				ServiceName:     service.Name,
				ServicePlanName: service.Plans[0].Name,
				Bindings:        instance.Bindings,
				Metadata:        instance.Metadata,
				AuditTrail:      instance.AuditTrail,
				State:           instance.State,
				OfferingId:      service.Id,
			}

			Convey("ConvertToApiServiceInstance should not return error and should return proper response", func() {
				result, err := ConvertToApiServiceInstance(service, instance)

				So(err, ShouldBeNil)
				So(result, ShouldResemble, expectedResult)
			})
		})
	})
}
