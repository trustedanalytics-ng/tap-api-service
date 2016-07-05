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
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHTTP "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *TapApiServiceApiOAuth2Connector) CreateApplicationInstance(blob multipart.File, manifest models.Manifest, timeout time.Duration) (catalogModels.Application, error) {
	connector := c.getApiOAuth2Connector("/applications")
	result := catalogModels.Application{}

	contentType, bodyBuf, err := c.prepareApplicationCreationForm(blob, manifest)
	if err != nil {
		logger.Error("Preparing application creation form failed: ", err)
		return result, err
	}

	req, err := http.NewRequest(http.MethodPost, connector.Url, bodyBuf)
	if err != nil {
		logger.Error("Preparing HTTP request failed: ", err)
		return result, err
	}
	req.Header.Add("Authorization", commonHTTP.GetOAuth2Header(connector.OAuth2))
	commonHTTP.SetContentType(req, contentType)

	logger.Infof("Doing: POST %v", connector.Url)

	var resp *http.Response
	wrapWithTemporaryTimeout(connector.Client, timeout, func() {
		resp, err = connector.Client.Do(req)
	})
	if err != nil {
		logger.Error("Make http request POST: ", err)
		return result, err
	}

	expectedStatusCode := http.StatusAccepted
	if resp.StatusCode != expectedStatusCode {
		return result, fmt.Errorf("wrong response code! expected: %v , got: %v", expectedStatusCode, resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Reading HTTP response body failed: ", err)
		return result, err
	}

	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) DeleteApplicationInstance(instanceId string) error {
	connector := c.getApiOAuth2Connector("/applications/%s", instanceId)
	_, err := commonHTTP.DeleteModel(connector, http.StatusNoContent)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) ListApplicationInstances() ([]models.ApplicationInstance, error) {
	connector := c.getApiOAuth2Connector("/applications")
	result := &[]models.ApplicationInstance{}
	_, err := commonHTTP.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) ScaleApplicationInstance(instanceId string, replication int) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/scale", instanceId)
	body := models.ScaleApplicationRequest{
		Replicas: replication,
	}
	result := &containerBrokerModels.MessageResponse{}
	_, err := commonHTTP.PutModel(connector, body, http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) StartApplicationInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/start", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := commonHTTP.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) StopApplicationInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/stop", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := commonHTTP.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) RestartApplicationInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/restart", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := commonHTTP.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) prepareApplicationCreationForm(blob multipart.File, manifest models.Manifest) (string, *bytes.Buffer, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	err := c.createBlobFormFile(blob, bodyWriter)
	if err != nil {
		return "", bodyBuf, err
	}
	c.createManifestFormFile(manifest, bodyWriter)
	if err != nil {
		return "", bodyBuf, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	return contentType, bodyBuf, nil
}

func (c *TapApiServiceApiOAuth2Connector) createBlobFormFile(blob multipart.File, bodyWriter *multipart.Writer) error {
	blobWriter, err := bodyWriter.CreateFormFile("blob", "blob.tar.gz")
	if err != nil {
		logger.Error("Error creating blob file field")
		return err
	}
	size, err := io.Copy(blobWriter, blob)
	if err != nil {
		logger.Error("Error copying blob to buffer")
		return err
	}
	logger.Infof("Written %v bytes of blob to buffer", size)
	return nil
}

func (c *TapApiServiceApiOAuth2Connector) createManifestFormFile(manifest models.Manifest,
	bodyWriter *multipart.Writer) error {

	manifestWriter, err := bodyWriter.CreateFormFile("manifest", "manifest.json")
	if err != nil {
		logger.Error("Error creating manifest file field")
		return err
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		logger.Error("Error marshalling manifest.json")
		return err
	}
	size, err := manifestWriter.Write(manifestBytes)
	if err != nil {
		logger.Error("Error writing manifest to buffer")
		return err
	}
	logger.Infof("Written %v bytes of manifest to buffer", size)
	return nil
}

func (c *TapApiServiceApiOAuth2Connector) DeleteApplication(instanceId string) error {
	connector := c.getApiOAuth2Connector("/applications/%s", instanceId)
	_, err := commonHTTP.DeleteModel(connector, http.StatusNoContent)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) GetApplicationInstance(applicationId string) (models.ApplicationInstance, error) {
	connector := c.getApiOAuth2Connector("/applications/%s", applicationId)
	result := &models.ApplicationInstance{}
	_, err := commonHTTP.GetModel(connector, http.StatusOK, result)
	return *result, err
}
