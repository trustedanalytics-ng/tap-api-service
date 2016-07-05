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

package main

import (
	"os"
	"sync"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-api-service/api"
	"github.com/trustedanalytics-ng/tap-api-service/models"
	uaaApi "github.com/trustedanalytics-ng/tap-api-service/uaa-connector"
	userManagementApi "github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	blobStoreApi "github.com/trustedanalytics-ng/tap-blob-store/client"
	catalogApi "github.com/trustedanalytics-ng/tap-catalog/client"
	containerBrokerApi "github.com/trustedanalytics-ng/tap-container-broker/client"
	httpGoCommon "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-go-common/util"
	imageFactoryApi "github.com/trustedanalytics-ng/tap-image-factory/client"
	templateRepositoryApi "github.com/trustedanalytics-ng/tap-template-repository/client"
)

var logger, _ = commonLogger.InitLogger("main")
var waitGroup = &sync.WaitGroup{}

func main() {
	go util.TerminationObserver(waitGroup, "Api-Service")

	initServices()

	router := setupRouter()

	httpGoCommon.StartServer(router)
}

func initServices() {
	templateRepositoryConnector, err := getTemplateRepositoryConnector()
	if err != nil {
		logger.Fatal("Can't connect with TAP-NG-template-provider!", err)
	}

	blobStoreConnector, err := getBlobStoreConnector()
	if err != nil {
		logger.Fatal("Can't connect with tap-blob-store! ", err)
	}

	catalogAPI, err := getCatalogAPI()
	if err != nil {
		logger.Fatal("Can't connect with TAP-NG-catalog!", err)
	}

	containerBrokerConnector, err := getContainerBrokerConnector()
	if err != nil {
		logger.Fatal("Can't connect with tap-container-broker! ", err)
	}

	imageFactoryConnector, err := getImageFactoryConnector()
	if err != nil {
		logger.Fatal("Can't connect with tap-image-factory! ", err)
	}

	uaaConnector, err := getUaaConnector()
	if err != nil {
		logger.Fatal("Can't connect with UAA! ", err)
	}

	userManagementConnectorFactory, err := getUserManagementConnectorFactory()
	if err != nil {
		logger.Fatal("Can't connect with user-management! ", err)
	}

	api.BrokerConfig = &api.Config{}
	api.BrokerConfig.TemplateRepositoryApi = templateRepositoryConnector
	api.BrokerConfig.CatalogApi = catalogAPI
	api.BrokerConfig.ContainerBrokerApi = containerBrokerConnector
	api.BrokerConfig.BlobStoreApi = blobStoreConnector
	api.BrokerConfig.ImageFactoryApi = imageFactoryConnector
	api.BrokerConfig.UaaApi = uaaConnector
	api.BrokerConfig.UserManagementApiFactory = userManagementConnectorFactory
}

func setupRouter() *web.Router {
	router := web.New(models.Context{})
	api.SetupRouter(router, true)
	return router
}

func getTemplateRepositoryConnector() (*templateRepositoryApi.TemplateRepositoryConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("TEMPLATE_REPOSITORY")
	if err != nil {
		panic(err.Error())
	}
	return templateRepositoryApi.NewTemplateRepositoryBasicAuth("https://"+address, username, password)
}

func getCatalogAPI() (catalogApi.TapCatalogApi, error) {

	address, username, password, err := util.GetConnectionParametersFromEnv("CATALOG")
	if err != nil {
		panic(err.Error())
	}
	return catalogApi.NewTapCatalogApiWithBasicAuth("https://"+address, username, password)
}

func getContainerBrokerConnector() (*containerBrokerApi.TapContainerBrokerApiConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("CONTAINER_BROKER")
	if err != nil {
		panic(err.Error())
	}
	return containerBrokerApi.NewTapContainerBrokerApiWithBasicAuth("https://"+address, username, password)
}

func getBlobStoreConnector() (*blobStoreApi.TapBlobStoreApiConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("BLOB_STORE")
	if err != nil {
		panic(err.Error())
	}
	return blobStoreApi.NewTapBlobStoreApiWithBasicAuth("https://"+address, username, password)
}

func getImageFactoryConnector() (*imageFactoryApi.TapImageFactoryApiConnector, error) {
	address, username, password, err := util.GetConnectionParametersFromEnv("IMAGE_FACTORY")
	if err != nil {
		panic(err.Error())
	}
	return imageFactoryApi.NewTapImageFactoryApiWithBasicAuth("https://"+address, username, password)
}

func getUaaConnector() (*uaaApi.UaaConnector, error) {
	client := os.Getenv("SSO_CLIENT")
	secret := os.Getenv("SSO_SECRET")
	return uaaApi.NewUaaBasicAuth(client, secret)
}

func getUserManagementConnectorFactory() (*userManagementApi.UserManagementApiConnectorFactory, error) {
	address, err := util.GetConnectionAddressFromEnvs("USER_MANAGEMENT")
	if err != nil {
		panic(err.Error())
	}
	return userManagementApi.NewUserManagementConnectorFactory(address)
}
