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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gocraft/web"

	apiServiceModels "github.com/trustedanalytics-ng/tap-api-service/models"
	containerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const (
	coreComponentDefaultProtocol = "http"
	defaultNamespaceName         = "default"
	coreComponentsDefaultPort    = "80"
	coreComponentsInfoEndpoint   = "%s://%s.%s:%s/info"
	timeoutForInfoEndpoint       = 10
)

var coreComponentsContainerVersions = []containerModels.VersionsResponse{}

func (c *Context) GetPlatformInfo(rw web.ResponseWriter, req *web.Request) {
	platformContext := apiServiceModels.PlatformInfo{}
	platformContext.ApiEndpoint = "https://api." + c.Domain
	platformContext.CliUrl = "https://api." + c.Domain + "/cli?ver=" + c.CliVersion
	platformContext.CliVersion = c.CliVersion
	platformContext.CoreOrganization = c.CoreOrganization
	platformContext.PlatformVersion = c.TapVersion
	platformContext.CdhVersion = c.CdhVersion
	platformContext.K8sVersion = c.K8sVersion
	visualisations_list := strings.Split(os.Getenv("VISUALISATIONS_LIST"), ",")
	for _, visualtisation := range visualisations_list {
		platformContext.ExternalTools.Visualizations = appendIfAvalaible(platformContext.ExternalTools.Visualizations, visualtisation)
	}

	commonHttp.WriteJson(rw, platformContext, http.StatusOK)
}

func (c *Context) GetPlatformComponents(rw web.ResponseWriter, req *web.Request) {
	platformComponents := fetchVersionsFromCoreComponents()
	commonHttp.WriteJson(rw, platformComponents, http.StatusOK)
}

func appendIfAvalaible(platformTools []apiServiceModels.PlatformTool, toolName string) (extendedPlatformTools []apiServiceModels.PlatformTool) {
	extendedPlatformTools = platformTools

	toolConfigMap, status, err := BrokerConfig.ContainerBrokerApi.GetConfigMap(toolName)
	if err != nil || status != http.StatusOK {
		logger.Warningf("Unable to get configMap for %s: %v", toolName, err)
		return
	}

	toolAvailability, err := strconv.ParseBool(toolConfigMap.Data["available"])
	if err != nil {
		logger.Warning("Unable to parse toolAvaliability")
		return
	}

	tool := apiServiceModels.PlatformTool{Name: toolName, Url: toolConfigMap.Data["uri"], Available: toolAvailability}
	extendedPlatformTools = append(platformTools, tool)
	return
}

func fetchVersionsFromCoreComponents() (versions []apiServiceModels.PlatformComponents) {
	versions = []apiServiceModels.PlatformComponents{}
	var err error
	var status int
	coreComponentsContainerVersions, status, err = BrokerConfig.ContainerBrokerApi.GetVersions()
	if err != nil || status != http.StatusOK {
		logger.Infof("Unable to get platform versions %s", err)
		return
	}

	var wg sync.WaitGroup
	responses := make(chan apiServiceModels.PlatformComponents, len(coreComponentsContainerVersions))

	for _, component := range coreComponentsContainerVersions {
		wg.Add(1)
		go getVersionForSpecificCoreApp(component, &wg, responses)
	}
	wg.Wait()
	close(responses)

	for response := range responses {
		versions = append(versions, response)
	}
	return versions
}

func getVersionForSpecificCoreApp(component containerModels.VersionsResponse, wg *sync.WaitGroup, ch chan apiServiceModels.PlatformComponents) {
	defer wg.Done()
	appVersion, err := getVersionFromInfoUrl(component.Name)
	if err != nil {
		logger.Infof("Endpoint /info for component: %s no yet implemented. Error: %s", component.Name, err.Error())
	} else {
		logger.Infof("Finished getting version info for: %s", component.Name)
	}
	extendedVersion := apiServiceModels.PlatformComponents{
		Name:          component.Name,
		Signature:     component.Signature,
		ImageVersions: component.ImageVersions,
		AppVersion:    appVersion,
	}
	ch <- extendedVersion
}

func getHttpClient() *http.Client {
	return &http.Client{Timeout: time.Duration(timeoutForInfoEndpoint * time.Second)}
}

var initHttpClient = getHttpClient

func getVersionFromInfoUrl(serviceName string) (string, error) {
	result := apiServiceModels.ComponentInfo{}
	coreComponentUrl := getCoreComponentInfoUrl(serviceName)
	response, err := initHttpClient().Get(coreComponentUrl)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if len(body) != 0 {
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", err
		}
	}
	return result.ComponentInfo.AppInfo.Version, nil
}

func getCoreComponentInfoUrl(serviceName string) string {
	return fmt.Sprintf(coreComponentsInfoEndpoint, coreComponentDefaultProtocol, serviceName, defaultNamespaceName, coreComponentsDefaultPort)
}
