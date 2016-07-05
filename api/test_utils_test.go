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
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gocraft/web"
	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics-ng/tap-api-service/client"
	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-api-service/uaa-connector"
	"github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	templateRepositoryModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	testDataDirPath     = "../testData"
	testApplicationsDir = "applications"
	blobFilename        = "tapng-sample-python-app.tar.gz"
	manifestFilename    = "manifest.json"

	apiPrefix = "v3"

	applicationID1           = "appID1"
	applicationName1         = "appName1"
	applicationImageAddress1 = "imageAddress1"

	serviceName1     = "serviceName1"
	serviceName2     = "serviceName2"
	serviceName3     = "serviceName3"
	wrongServiceName = "wrongServiceName"

	serviceID1     = "serviceID1"
	serviceID2     = "serviceID2"
	serviceID3     = "serviceID3"
	wrongServiceID = "wrongServiceID"

	serviceDescription1 = "serviceDescription1"
	serviceBindable1    = true
	serviceState1       = catalogModels.ServiceStateDeploying
	serviceTemplateID1  = "serviceTemplateID1"

	instanceID1     = "instanceID1"
	instanceID2     = "instanceID2"
	instanceID3     = "instanceID3"
	instanceID4     = "instanceID4"
	instanceID5     = "instanceID5"
	wrongInstanceID = "wrongInstanceID"

	instanceName1     = "instanceName1"
	instanceName2     = "instanceName2"
	instanceName3     = "instanceName3"
	instanceName4     = "instanceName4"
	wrongInstanceName = "wrongInstanceName"

	planName1     = "plan1"
	planName2     = "plan2"
	wrongPlanName = "wrongPlanName"

	planID1 = "planID1"
	planID2 = "planID2"

	planDescription1 = "planDescription1"
	planDescription2 = "planDescription2"

	planCost1 = "free"
	planCost2 = "very expensive"

	metadataID1    = "metadataID1"
	metadataID2    = "metadataID2"
	metadataValue1 = "metadataValue1"
	metadataValue2 = "metadataValue2"
)

type createOfferingMockStructures struct {
	newCatalogTemplate             catalogModels.Template
	catalogTemplateWithId          catalogModels.Template
	newTemplate                    templateRepositoryModels.RawTemplate
	patchUpdateTemplateState       catalogModels.Patch
	catalogTemplateInStateReady    catalogModels.Template
	newCatalogService              catalogModels.Service
	catalogServiceWithId           catalogModels.Service
	patchUpdateCatalogServiceState catalogModels.Patch
	catalogServiceInStateReady     catalogModels.Service
}

type mocksAndRouter struct {
	router                       *web.Router
	catalogApiMock               *MockTapCatalogApi
	blobStoreApiMock             *MockTapBlobStoreApi
	containerBrokerApiMock       *MockTapContainerBrokerApi
	uaaApiMock                   *uaa_connector.MockUaaApi
	userManagementApiFactoryMock *user_management_connector.MockUserManagementFactory
	userManagementApiMock        *user_management_connector.MockUserManagementApi
	templateRepositoryApiMock    *MockTemplateRepository
	mockCtrl                     *gomock.Controller
}

func prepareMocksAndRouter(t *testing.T) mocksAndRouter {
	return prepareMocksAndRouterGeneric(t, false)
}

func prepareMocksAndRouterWithClient(t *testing.T) (mocksAndRouter, client.TapApiServiceApi) {
	mocksConfig := prepareMocksAndRouterGeneric(t, false)
	testServer := httptest.NewServer(mocksConfig.router)

	mockedClient, err := client.NewTapApiServiceApiWithOAuth2(testServer.URL, "tokenType", "token")
	if err != nil {
		t.Fatalf("initializing mocked client error: %v", err)
	}

	return mocksConfig, mockedClient
}

func prepareMocksAndRouterWithOauth2Activated(t *testing.T) mocksAndRouter {
	return prepareMocksAndRouterGeneric(t, true)
}

func prepareMocksAndRouterGeneric(t *testing.T, oauthMiddlewareActivated bool) mocksAndRouter {
	mockCtrl := gomock.NewController(t)

	router := web.New(models.Context{})
	SetupRouter(router, oauthMiddlewareActivated)

	result := mocksAndRouter{
		router:                       router,
		catalogApiMock:               NewMockTapCatalogApi(mockCtrl),
		blobStoreApiMock:             NewMockTapBlobStoreApi(mockCtrl),
		containerBrokerApiMock:       NewMockTapContainerBrokerApi(mockCtrl),
		uaaApiMock:                   uaa_connector.NewMockUaaApi(mockCtrl),
		userManagementApiFactoryMock: user_management_connector.NewMockUserManagementFactory(mockCtrl),
		userManagementApiMock:        user_management_connector.NewMockUserManagementApi(mockCtrl),
		templateRepositoryApiMock:    NewMockTemplateRepository(mockCtrl),
		mockCtrl:                     mockCtrl,
	}

	BrokerConfig = &Config{
		CatalogApi:         result.catalogApiMock,
		BlobStoreApi:       result.blobStoreApiMock,
		ContainerBrokerApi: result.containerBrokerApiMock,
		UaaApi:             result.uaaApiMock,
		UserManagementApiFactory: result.userManagementApiFactoryMock,
		TemplateRepositoryApi:    result.templateRepositoryApiMock,
	}

	return result
}

func prepareMocks(t *testing.T) *MockTapCatalogApi {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	BrokerConfig = &Config{}
	catalogApiMock := NewMockTapCatalogApi(mockCtrl)
	BrokerConfig.CatalogApi = catalogApiMock

	return catalogApiMock
}

func SendForm(path string, body io.Reader, contentType string, r *web.Router) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", path, body)
	req.Header.Set("Content-Type", contentType)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func SendGet(path string, r *web.Router) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", path, nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func PrepareCreateApplicationForm(blobFilename string, manifestFilename string) (bodyBuf *bytes.Buffer, contentType string) {
	bodyBuf = &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("blob", "blob.tar.gz")
	checkTestingError(err)
	fh, err := os.Open(blobFilename)
	checkTestingError(err)
	io.Copy(fileWriter, fh)
	fh.Close()

	fileWriter, err = bodyWriter.CreateFormFile("manifest", "manifest.json")
	checkTestingError(err)
	fhandler2, err := os.Open(manifestFilename)
	checkTestingError(err)
	io.Copy(fileWriter, fhandler2)
	fhandler2.Close()

	contentType = bodyWriter.FormDataContentType()
	bodyWriter.Close()

	return
}

func readAndAssertJson(rr *httptest.ResponseRecorder, retstruct interface{}) {
	err := json.Unmarshal(rr.Body.Bytes(), &retstruct)
	So(err, ShouldBeNil)
}

func checkTestingError(err error) {
	if err != nil {
		panic(err)
	}
}

func addToQuery(link, name, value string) string {
	if value == "" {
		return link
	}
	res := link

	if !strings.Contains(link, "?") {
		res += "?"
	} else {
		res += "&"
	}

	res += fmt.Sprintf("%s=%s", name, value)

	return res
}
