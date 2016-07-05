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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

type TapApiImageFactoryApi interface {
	BuildImage(imageId string) error
}

func NewTapImageFactoryApiWithBasicAuth(address, username, password string) (*TapImageFactoryApiConnector, error) {
	client, _, err := brokerHttp.GetHttpClient()
	if err != nil {
		return nil, err
	}
	return &TapImageFactoryApiConnector{address, username, password, client}, nil
}

type TapImageFactoryApiConnector struct {
	Address  string
	Username string
	Password string
	Client   *http.Client
}

func (c *TapImageFactoryApiConnector) getApiConnector(url string) brokerHttp.ApiConnector {
	return brokerHttp.ApiConnector{
		BasicAuth: &brokerHttp.BasicAuth{User: c.Username, Password: c.Password},
		Client:    c.Client,
		Url:       url,
	}
}

type RequestJson struct {
	ImageId string `json:"id"`
}

func (c *TapImageFactoryApiConnector) BuildImage(imageId string) error {
	req_json := RequestJson{
		ImageId: imageId,
	}
	requestBodyByte, err := json.Marshal(req_json)
	if err != nil {
		return err
	}
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/image", c.Address))
	status, _, err := brokerHttp.RestPOST(connector.Url, string(requestBodyByte), brokerHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	if err != nil || status != http.StatusAccepted {
		return errors.New(fmt.Sprintf("Error building image. Responded with %v. Error: %v", status, err))
	}
	return nil
}

func (c *TapImageFactoryApiConnector) GetImageFactoryHealth() error {
	connector := c.getApiConnector(fmt.Sprintf("%s/api/v1/healthz", c.Address))
	status, _, err := brokerHttp.RestGET(connector.Url, brokerHttp.GetBasicAuthHeader(connector.BasicAuth), connector.Client)
	if status != http.StatusOK {
		err = errors.New("Invalid health status: " + string(status))
	}
	return err
}
