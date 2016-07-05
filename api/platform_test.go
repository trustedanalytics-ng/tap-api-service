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
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"
	apiServiceModels "github.com/trustedanalytics-ng/tap-api-service/models"
	containerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	"net/url"
)

const (
	fakeAppVersion               = "0.0.1"
	fakeResponseFromInfoEndpoint = `{
					    "info": {
					        "app":
					        {
					           "id": "genericCoreApp",
					           "version" : "` + fakeAppVersion + `"
					        }
					    }
					}`
)

func TestFetchVersionsFromCoreComponents(t *testing.T) {
	mocksAndRouter := prepareMocksAndRouter(t)
	server := initMockClient()
	defer server.Close()

	Convey("For working /info endpoint fetchVersionsFromCoreComponents should return proper responses", t, func() {
		Convey("Given proper response from conatiner-broker, should return valid response", func() {
			fakeVersions := getFakeVersionsFromContainerBroker()
			mocksAndRouter.containerBrokerApiMock.EXPECT().GetVersions().Return(fakeVersions, http.StatusOK, nil)
			actualResult := fetchVersionsFromCoreComponents()
			expectedResult := getValidVersionResponse()

			So(len(expectedResult), ShouldEqual, len(actualResult))
			for _, expected := range expectedResult {
				for _, actual := range actualResult {
					if expected == actual {
						So(actual, ShouldResemble, expected)
					}
				}
			}
		})
	})
}

func initMockClient() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, fakeResponseFromInfoEndpoint)
	}))
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	initHttpClient = func() *http.Client { return &http.Client{Transport: transport} }
	return server
}

func getValidVersionResponse() []apiServiceModels.PlatformComponents {
	return []apiServiceModels.PlatformComponents{
		{Name: "console", ImageVersions: "0.8.0.2253", Signature: "dcb25e66ea5e6ab93c07e6c428a3529497954a4991e873feaacd3d1e92eee6a4", AppVersion: fakeAppVersion},
		{Name: "catalog", ImageVersions: "0.8.0.2253", Signature: "f55ed1aa03ef29b72bc7ef632ba47a7e38348c3422a8df18503f4813f5e4e72e", AppVersion: fakeAppVersion},
		{Name: "blob-store", ImageVersions: "0.8.0.2253", Signature: "bc2e23fba06546db8a121a7c892f04977111063e90cd2c8097d6f8cfa8b85268", AppVersion: fakeAppVersion},
		{Name: "auth-gateway", ImageVersions: "0.8.0.2253", Signature: "0a705c5cb9b9d1f738ab8c6d3d2121692130d2e9cf5ce8a3a9951e29ca2bbaff", AppVersion: fakeAppVersion},
	}
}

func getFakeVersionsFromContainerBroker() []containerModels.VersionsResponse {
	return []containerModels.VersionsResponse{
		{Name: "console", Image: "127.0.0.1:30000/console:0.8.0.2253", ImageVersions: "0.8.0.2253", Signature: "dcb25e66ea5e6ab93c07e6c428a3529497954a4991e873feaacd3d1e92eee6a4"},
		{Name: "catalog", Image: "127.0.0.1:30000/tap-catalog:0.8.0.2253", ImageVersions: "0.8.0.2253", Signature: "f55ed1aa03ef29b72bc7ef632ba47a7e38348c3422a8df18503f4813f5e4e72e"},
		{Name: "blob-store", Image: "127.0.0.1:30000/tap-blob-store:0.8.0.2253", ImageVersions: "0.8.0.2253", Signature: "bc2e23fba06546db8a121a7c892f04977111063e90cd2c8097d6f8cfa8b85268"},
		{Name: "auth-gateway", Image: "127.0.0.1:30000/auth-gateway:0.8.0.2253", ImageVersions: "0.8.0.2253", Signature: "0a705c5cb9b9d1f738ab8c6d3d2121692130d2e9cf5ce8a3a9951e29ca2bbaff"},
	}
}
