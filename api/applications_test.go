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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/twinj/uuid"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func getFakeServicesInstances() []catalogModels.Instance {
	return []catalogModels.Instance{
		{Id: serviceID1, Name: serviceName1},
		{Id: serviceID2, Name: serviceName2},
		{Id: serviceID3, Name: serviceName3},
	}
}

func TestGetInstanceDependencyForBindings(t *testing.T) {
	catalogApiMock := prepareMocks(t)
	fakeServicesInstances := getFakeServicesInstances()

	Convey(fmt.Sprintf("For list of services instances %v", fakeServicesInstances), t, func() {
		catalogApiMock.EXPECT().ListServicesInstances().Return(fakeServicesInstances, http.StatusOK, nil)

		Convey("getInstanceDependencyForBindings should return proper response for set of test cases", func() {
			testCases := []struct {
				input   []string
				isError bool
				output  []catalogModels.InstanceDependency
			}{
				{input: []string{serviceName1, serviceName2, serviceName3}, isError: false,
					output: []catalogModels.InstanceDependency{
						{Id: serviceID1},
						{Id: serviceID2},
						{Id: serviceID3},
					},
				},
				{input: []string{serviceName2}, isError: false, output: []catalogModels.InstanceDependency{{Id: serviceID2}}},
				{input: []string{}, isError: false, output: []catalogModels.InstanceDependency{}},
				{input: []string{wrongServiceName}, isError: true, output: []catalogModels.InstanceDependency{}},
			}

			for _, tc := range testCases {
				Convey(fmt.Sprintf("For input bindings %v it should return proper response", tc.input), func() {
					output, err := getInstanceDependencyForBindings(tc.input)
					if tc.isError {
						Convey("error should not be nil", func() {
							So(err, ShouldNotBeNil)
						})
					} else {
						Convey("error should be nil", func() {
							So(err, ShouldBeNil)
						})
						Convey("InstanceDependency array should be proper", func() {
							So(output, ShouldResemble, tc.output)
						})
					}
				})
			}
		})
	})
}

func TestCreateApplicationInstance(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)

	applicationId,
		imageId,
		fakeServicesInstances,
		application,
		fakeApplicationResponse,
		fakeApplicationResponseFromUpdate,
		applicationImageIDPatch,
		imageStatePatch,
		image,
		properResponse := arrangeForTestCreateApplicationInstance()

	Convey(fmt.Sprint("Given blob, manifest and all structs handled by CreateApplicationInstance"), t, func() {
		Convey("All inner functions should be invoked properly", func() {
			mocksAndRouter.containerBrokerApiMock.EXPECT().GetVersions().Return([]containerBrokerModels.VersionsResponse{}, http.StatusOK, nil)
			mocksAndRouter.catalogApiMock.EXPECT().ListServicesInstances().
				Return(fakeServicesInstances, http.StatusOK, nil)
			mocksAndRouter.catalogApiMock.EXPECT().AddApplication(application).
				Return(fakeApplicationResponse, http.StatusAccepted, nil)
			mocksAndRouter.catalogApiMock.EXPECT().UpdateApplication(applicationId, []catalogModels.Patch{applicationImageIDPatch}).
				Return(fakeApplicationResponseFromUpdate, http.StatusOK, nil)
			mocksAndRouter.blobStoreApiMock.EXPECT().StoreBlob(imageId, gomock.Any()).Return(nil)
			mocksAndRouter.catalogApiMock.EXPECT().AddImage(image).Return(image, http.StatusAccepted, nil)
			mocksAndRouter.catalogApiMock.EXPECT().UpdateImage(imageId, []catalogModels.Patch{imageStatePatch}).Return(image, http.StatusOK, nil)

			bodyBuf, contentType := PrepareCreateApplicationForm(
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, blobFilename),
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, manifestFilename))

			rr := SendForm(fmt.Sprintf("/api/%s/%s", apiPrefix, "applications"), bodyBuf, contentType, mocksAndRouter.router)
			response := catalogModels.Application{}

			Convey("/api/v3/applications should return proper status code", func() {
				So(rr.Code, ShouldEqual, http.StatusAccepted)
			})

			Convey("Response json should properly unmarshal", func() {
				readAndAssertJson(rr, &response)

				Convey("real response should resemble proper response", func() {
					So(response, ShouldResemble, properResponse)
				})
			})
		})

		Convey("Blob and manifest are required", func() {
			reqBody, err := json.Marshal("")
			So(err, ShouldBeNil)

			response := commonHttp.SendRequest("POST", "/api/v3/applications", reqBody, mocksAndRouter.router, t)
			So(response.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("Application name cannot equal to any core component", func() {
			mocksAndRouter.containerBrokerApiMock.EXPECT().GetVersions().Return([]containerBrokerModels.VersionsResponse{
				{Name: application.Name}}, http.StatusOK, nil).Times(1)

			bodyBuf, contentType := PrepareCreateApplicationForm(
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, blobFilename),
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, manifestFilename))

			rr := SendForm(fmt.Sprintf("/api/%s/%s", apiPrefix, "applications"), bodyBuf, contentType, mocksAndRouter.router)

			So(rr.Code, ShouldEqual, http.StatusBadRequest)
		})
		Convey("Core components container versions should be get from cache, GetVersions() should not be called", func() {
			mocksAndRouter.containerBrokerApiMock.EXPECT().GetVersions().Return([]containerBrokerModels.VersionsResponse{
				{Name: application.Name}}, http.StatusOK, nil).Times(0)

			bodyBuf, contentType := PrepareCreateApplicationForm(
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, blobFilename),
				fmt.Sprintf("%s/%s/%s", testDataDirPath, testApplicationsDir, manifestFilename))

			rr := SendForm(fmt.Sprintf("/api/%s/%s", apiPrefix, "applications"), bodyBuf, contentType, mocksAndRouter.router)

			So(rr.Code, ShouldEqual, http.StatusBadRequest)
		})
		Reset(func() {
			mocksAndRouter.mockCtrl.Finish()
		})
	})
}

func arrangeForTestCreateApplicationInstance() (applicationId string,
	imageId string,
	fakeServicesInstances []catalogModels.Instance,
	applicationEntity catalogModels.Application,
	fakeApplicationResponse catalogModels.Application,
	fakeApplicationResponseFromUpdate catalogModels.Application,
	applicationImageIDPatch catalogModels.Patch,
	imageStatePatch catalogModels.Patch,
	image catalogModels.Image,
	properResponse catalogModels.Application) {
	manifestBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/%s",
		testDataDirPath, testApplicationsDir, manifestFilename))
	checkTestingError(err)

	manifest := models.Manifest{}
	err = json.Unmarshal(manifestBytes, &manifest)
	checkTestingError(err)

	fakeServicesInstances = []catalogModels.Instance{
		{Id: uuid.NewV4().String()},
		{Id: uuid.NewV4().String()},
		{Id: uuid.NewV4().String()},
	}

	applicationEntity = catalogModels.Application{
		Name:                 manifest.Name,
		Replication:          manifest.Instances,
		TemplateId:           genericApplicationTemplateID,
		InstanceDependencies: []catalogModels.InstanceDependency{},
	}

	applicationId = uuid.NewV4().String()
	imageId = catalogModels.GenerateImageId(applicationId)

	fakeApplicationResponse = catalogModels.Application{
		Name:        manifest.Name,
		Replication: manifest.Instances,
		TemplateId:  genericApplicationTemplateID,
		Id:          applicationId,
	}

	fakeApplicationResponseFromUpdate = catalogModels.Application{
		Name:        manifest.Name,
		Replication: manifest.Instances,
		TemplateId:  genericApplicationTemplateID,
		Id:          applicationId,
		ImageId:     imageId,
	}

	applicationImageIDPatch, err = builder.MakePatch("ImageId", imageId, catalogModels.OperationUpdate)

	image = catalogModels.Image{
		Id:       imageId,
		Type:     manifest.ImageType,
		BlobType: catalogModels.BlobTypeTarGz,
	}

	imageStatePatch, err = builder.MakePatch("State", catalogModels.ImageStatePending, catalogModels.OperationUpdate)

	properResponse = catalogModels.Application{
		Name:        manifest.Name,
		Replication: manifest.Instances,
		TemplateId:  genericApplicationTemplateID,
		Id:          applicationId,
		ImageId:     imageId,
	}

	return
}

func TestUpdateImageState(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)

	fakeAppId := uuid.NewV4().String()
	imageId := catalogModels.GenerateImageId(fakeAppId)
	username := "fake-username"
	newState := catalogModels.ImageStatePending

	fakeErr := errors.New("update failed!")

	expectedPatch, _ := builder.MakePatch("State", newState, catalogModels.OperationUpdate)
	expectedPatch.Username = username

	expectedImage := catalogModels.Image{
		State: newState,
	}

	Convey(fmt.Sprint("Given imageId of image which state should be updated to newState"), t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().UpdateImage(imageId, []catalogModels.Patch{expectedPatch}).Return(expectedImage, http.StatusOK, nil)

		realImage, realErr := updateImageState(imageId, newState, username)

		Convey("updateImageState should return Image with updated state", func() {
			So(realImage, ShouldResemble, expectedImage)
		})
		Convey("updateImageState should not return any error", func() {
			So(realErr, ShouldBeNil)
		})
	})

	Convey(fmt.Sprint("Given bad imageId of image which state should be updated to newState"), t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().UpdateImage(imageId, []catalogModels.Patch{expectedPatch}).Return(expectedImage, http.StatusNotFound, errors.New("update failed!"))

		realImage, realErr := updateImageState(imageId, newState, username)

		Convey("updateImageState should return Image with updated state", func() {
			So(realImage, ShouldResemble, expectedImage)
		})
		Convey("updateImageState should return an error", func() {
			So(realErr, ShouldResemble, fakeErr)
		})
	})
}

func TestGetApplicationInstance(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)

	fakeAppID := uuid.NewV4().String()
	fakeInstanceID := uuid.NewV4().String()
	fakeApp := catalogModels.Application{
		Id:                   fakeInstanceID,
		Name:                 "some-name",
		Description:          "some-desc",
		ImageId:              catalogModels.GenerateImageId(fakeAppID),
		Replication:          1,
		TemplateId:           uuid.NewV4().String(),
		InstanceDependencies: []catalogModels.InstanceDependency{},
	}

	fakeInstance := catalogModels.Instance{
		Id:      fakeInstanceID,
		ClassId: fakeApp.Id,
		Type:    catalogModels.InstanceTypeApplication,
	}

	fakeImage := catalogModels.Image{
		Id: catalogModels.GenerateImageId(fakeAppID),
	}

	expectedAppInstance := models.ApplicationInstance{
		Id:               fakeInstanceID,
		Name:             fakeApp.Name,
		Type:             catalogModels.InstanceTypeApplication,
		State:            waitingForImageApplicationInstanceState,
		Replication:      fakeApp.Replication,
		ImageState:       fakeImage.State,
		ImageType:        fakeImage.Type,
		Memory:           apiApplicationInstanceMemoryDefault,
		DiskQuota:        apiApplicationInstanceDiskQuotaDefault,
		RunningInstances: fakeApp.Replication,
		Urls:             []string{""},
	}

	Convey("Given app, its instance and image", t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().ListApplicationInstances(fakeAppID).Return([]catalogModels.Instance{fakeInstance}, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetApplication(fakeAppID).Return(fakeApp, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetImage(fakeApp.ImageId).Return(fakeImage, http.StatusOK, nil)

		Convey("When the getApplicationInstance(appID) is invoked", func() {
			realAppInstance, err := getApplicationInstance(fakeAppID)

			Convey("it shouldn't return any error", func() {
				So(err, ShouldBeNil)
			})
			Convey("it should return expected api-service application", func() {
				So(realAppInstance, ShouldResemble, expectedAppInstance)
			})
		})
	})

	Convey("Given nonexistent app", t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().ListApplicationInstances(fakeAppID).Return([]catalogModels.Instance{},
			http.StatusNotFound,
			errors.New("Bad response status: 404, expected status was: 200. Response body: Key not found: /Instances"))

		Convey("When the getApplicationInstance(appID) is invoked", func() {
			_, err := getApplicationInstance(fakeAppID)

			Convey("getApplicationInstance should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given appId; Catalog failing during /Application fetch", t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().ListApplicationInstances(fakeAppID).Return([]catalogModels.Instance{fakeInstance}, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetApplication(fakeAppID).Return(fakeApp, http.StatusInternalServerError, errors.New("some problem"))

		Convey("When the getApplicationInstance(appID) is invoked", func() {
			_, err := getApplicationInstance(fakeAppID)

			Convey("getApplicationInstance should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given appId; ParseToApiApplication failing", t, func() {
		mocksAndRouter.catalogApiMock.EXPECT().ListApplicationInstances(fakeAppID).Return([]catalogModels.Instance{fakeInstance}, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetApplication(fakeAppID).Return(fakeApp, http.StatusOK, nil)
		mocksAndRouter.catalogApiMock.EXPECT().GetImage(fakeApp.ImageId).Return(fakeImage, http.StatusInternalServerError, errors.New("some error"))

		Convey("When the getApplicationInstance(appID) is invoked", func() {
			_, err := getApplicationInstance(fakeAppID)

			Convey("getApplicationInstance should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
