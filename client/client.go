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
	"fmt"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	userManagement "github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	brokerHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
)

const (
	apiVer                        = "v3"
	DefaultPushApplicationTimeout = 0
	defaultGetLogsTimeout         = 1 * time.Minute
)

var logger, _ = commonLogger.InitLogger("client")

type TapApiServiceApi interface {
	GetPlatformInfo() (models.PlatformInfo, error)

	GetApplicationBindings(applicationId string) (models.InstanceBindings, error)
	GetServiceBindings(serviceId string) (models.InstanceBindings, error)
	BindToApplicationInstance(bindingRequest models.InstanceBindingRequest, applicationId string) (containerBrokerModels.MessageResponse, error)
	BindToServiceInstance(bindingRequest models.InstanceBindingRequest, serviceId string) (containerBrokerModels.MessageResponse, error)
	UnbindServiceFromApplicationInstance(serviceId, applicationId string) (int, error)
	UnbindApplicationFromApplicationInstance(srcApplicationId, dstApplicationId string) (int, error)
	UnbindServiceFromServiceInstance(srcServiceId, dstApplicationId string) (int, error)
	UnbindApplicationFromServiceInstance(applicationId, serviceId string) (int, error)

	CreateApplicationInstance(blob multipart.File, manifest models.Manifest, timeout time.Duration) (catalogModels.Application, error)
	CreateOffer(serviceWithTemplate models.ServiceDeploy) ([]catalogModels.Service, error)
	CreateServiceInstance(instance models.ServiceInstanceRequest) (containerBrokerModels.MessageResponse, error)

	DeleteOffering(serviceId string) error
	DeleteServiceInstance(instanceId string) error
	DeleteApplicationInstance(instanceId string) error

	GetOfferings() ([]models.Offering, error)
	GetOffering(offeringId string) (models.Offering, error)
	GetApplicationInstance(applicationId string) (models.ApplicationInstance, error)
	GetServiceInstance(serviceId string) (models.ServiceInstance, error)
	GetApplicationLogs(applicationId string) (map[string]string, error)
	GetServiceLogs(serviceId string) (map[string]string, error)
	GetInstanceCredentials(instanceId string) ([]containerBrokerModels.ContainerCredenials, error)

	ListApplicationInstances() ([]models.ApplicationInstance, error)
	ListServiceInstances() ([]models.ServiceInstance, error)

	StartApplicationInstance(applicationId string) (containerBrokerModels.MessageResponse, error)
	StopApplicationInstance(applicationId string) (containerBrokerModels.MessageResponse, error)
	RestartApplicationInstance(applicationId string) (containerBrokerModels.MessageResponse, error)
	ScaleApplicationInstance(applicationId string, replication int) (containerBrokerModels.MessageResponse, error)

	StartServiceInstance(serviceId string) (containerBrokerModels.MessageResponse, error)
	StopServiceInstance(serviceId string) (containerBrokerModels.MessageResponse, error)
	RestartServiceInstance(serviceId string) (containerBrokerModels.MessageResponse, error)

	GetInvitations() ([]string, error)
	SendInvitation(email string) (userManagement.InvitationResponse, error)
	ResendInvitation(email string) error
	DeleteInvitation(email string) error

	ExposeService(serviceId string, exposed bool) ([]string, int, error)

	GetUsers() ([]userManagement.UaaUser, error)
	ChangeCurrentUserPassword(password, newPassword string) error
	DeleteUser(email string) error
}

type TapApiServiceApiOAuth2Connector struct {
	Address   string
	TokenType string
	Token     string
	Client    *http.Client
}

func SetLoggerLevel(level string) error {
	return commonLogger.SetLoggerLevel(logger, level)
}

func NewTapApiServiceApiWithOAuth2(address, tokenType, token string) (TapApiServiceApi, error) {
	return NewTapApiServiceApiWithOAuth2AndCustomSSLValidation(address, tokenType, token, false)
}

func NewTapApiServiceApiWithOAuth2AndCustomSSLValidation(address, tokenType, token string, skipSSLValidation bool) (tapApiServiceApi TapApiServiceApi, err error) {
	client, _, err := brokerHttp.GetHttpClientWithCustomSSLValidation(skipSSLValidation)
	if err != nil {
		return
	}

	tapApiServiceApi = &TapApiServiceApiOAuth2Connector{
		Address:   address,
		TokenType: tokenType,
		Token:     token,
		Client:    client,
	}
	return
}

func getAddressCommon(addr string, format string, args ...interface{}) string {
	return fmt.Sprintf("%s/api/%s", addr, apiVer) + fmt.Sprintf(format, args...)
}

func wrapWithTemporaryTimeout(client *http.Client, tempTimeout time.Duration, callbackAfterSettingTimeout func()) {
	oldTimeout := client.Timeout
	setClientTimeout(client, tempTimeout)
	defer setClientTimeout(client, oldTimeout)
	callbackAfterSettingTimeout()
}

func setClientTimeout(client *http.Client, timeout time.Duration) {
	client.Timeout = timeout
}

func (c *TapApiServiceApiOAuth2Connector) getAddress(endpointFormat string, args ...interface{}) string {
	return getAddressCommon(c.Address, endpointFormat, args...)
}

func (c *TapApiServiceApiOAuth2Connector) getApiOAuth2Connector(endpointFormat string, args ...interface{}) brokerHttp.ApiConnector {
	return brokerHttp.ApiConnector{
		OAuth2: &brokerHttp.OAuth2{TokenType: c.TokenType, Token: c.Token},
		Client: c.Client,
		Url:    c.getAddress(endpointFormat, args...),
	}
}

func (c *TapApiServiceApiOAuth2Connector) CreateServiceInstance(instance models.ServiceInstanceRequest) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/services")
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PostModel(connector, instance, http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) CreateOffer(serviceWithTemplate models.ServiceDeploy) ([]catalogModels.Service, error) {
	connector := c.getApiOAuth2Connector("/offerings")
	result := &[]catalogModels.Service{}
	_, err := brokerHttp.PostModel(connector, serviceWithTemplate, http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) DeleteOffering(serviceId string) error {
	connector := c.getApiOAuth2Connector("/offerings/%s", serviceId)
	_, err := brokerHttp.DeleteModel(connector, http.StatusAccepted)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) DeleteServiceInstance(instanceId string) error {
	connector := c.getApiOAuth2Connector("/services/%s", instanceId)
	_, err := brokerHttp.DeleteModel(connector, http.StatusAccepted)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) GetPlatformInfo() (models.PlatformInfo, error) {
	connector := c.getApiOAuth2Connector("/platform_info")
	result := &models.PlatformInfo{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetOfferings() ([]models.Offering, error) {
	connector := c.getApiOAuth2Connector("/offerings")
	result := &[]models.Offering{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetOffering(offeringId string) (models.Offering, error) {
	connector := c.getApiOAuth2Connector("/offerings/%s", offeringId)
	result := &models.Offering{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetApplicationLogs(applicationId string) (map[string]string, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/logs", applicationId)

	result := make(map[string]string)
	var err error
	wrapWithTemporaryTimeout(connector.Client, defaultGetLogsTimeout, func() {
		_, err = brokerHttp.GetModel(connector, http.StatusOK, &result)
	})
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetServiceLogs(serviceId string) (map[string]string, error) {
	connector := c.getApiOAuth2Connector("/services/%s/logs", serviceId)

	result := make(map[string]string)
	var err error
	wrapWithTemporaryTimeout(connector.Client, defaultGetLogsTimeout, func() {
		_, err = brokerHttp.GetModel(connector, http.StatusOK, &result)
	})
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetInstanceCredentials(instanceId string) ([]containerBrokerModels.ContainerCredenials, error) {
	connector := c.getApiOAuth2Connector("/services/%s/credentials", instanceId)
	result := []containerBrokerModels.ContainerCredenials{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) ListServiceInstances() ([]models.ServiceInstance, error) {
	connector := c.getApiOAuth2Connector("/services")
	result := &[]models.ServiceInstance{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetServiceInstance(serviceId string) (models.ServiceInstance, error) {
	connector := c.getApiOAuth2Connector("/services/%s", serviceId)
	result := &models.ServiceInstance{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) StartServiceInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/services/%s/start", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) StopServiceInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/services/%s/stop", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) RestartServiceInstance(instanceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/services/%s/restart", instanceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PutModel(connector, "", http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetApplicationBindings(applicationId string) (models.InstanceBindings, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/bindings", applicationId)
	result := &models.InstanceBindings{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetServiceBindings(serviceId string) (models.InstanceBindings, error) {
	connector := c.getApiOAuth2Connector("/services/%s/bindings", serviceId)
	result := &models.InstanceBindings{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) BindToApplicationInstance(bindingRequest models.InstanceBindingRequest, applicationId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/bindings", applicationId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PostModel(connector, bindingRequest, http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) BindToServiceInstance(bindingRequest models.InstanceBindingRequest, serviceId string) (containerBrokerModels.MessageResponse, error) {
	connector := c.getApiOAuth2Connector("/services/%s/bindings", serviceId)
	result := &containerBrokerModels.MessageResponse{}
	_, err := brokerHttp.PostModel(connector, bindingRequest, http.StatusAccepted, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) UnbindServiceFromApplicationInstance(serviceId, applicationId string) (int, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/bindings/services/%s", applicationId, serviceId)
	return brokerHttp.DeleteModelWithBody(connector, "", http.StatusAccepted)
}

func (c *TapApiServiceApiOAuth2Connector) UnbindApplicationFromApplicationInstance(srcApplicationId, dstApplicationId string) (int, error) {
	connector := c.getApiOAuth2Connector("/applications/%s/bindings/applications/%s", dstApplicationId, srcApplicationId)
	return brokerHttp.DeleteModelWithBody(connector, "", http.StatusAccepted)
}

func (c *TapApiServiceApiOAuth2Connector) UnbindServiceFromServiceInstance(srcServiceId, dstServiceId string) (int, error) {
	connector := c.getApiOAuth2Connector("/services/%s/bindings/services/%s", dstServiceId, srcServiceId)
	return brokerHttp.DeleteModelWithBody(connector, "", http.StatusAccepted)
}

func (c *TapApiServiceApiOAuth2Connector) UnbindApplicationFromServiceInstance(applicationId, serviceId string) (int, error) {
	connector := c.getApiOAuth2Connector("/services/%s/bindings/applications/%s", serviceId, applicationId)
	return brokerHttp.DeleteModelWithBody(connector, "", http.StatusAccepted)
}

func (c *TapApiServiceApiOAuth2Connector) SendInvitation(email string) (userManagement.InvitationResponse, error) {
	connector := c.getApiOAuth2Connector("/users/invitations")
	body := userManagement.InvitationRequest{
		Email: email,
	}
	result := &userManagement.InvitationResponse{}
	_, err := brokerHttp.PostModel(connector, body, http.StatusCreated, result)
	return *result, err
}

func (c *TapApiServiceApiOAuth2Connector) ResendInvitation(email string) error {
	connector := c.getApiOAuth2Connector("/users/invitations/resend")
	body := userManagement.InvitationRequest{
		Email: email,
	}
	_, err := brokerHttp.PostModel(connector, body, http.StatusCreated, "")
	return err
}

func (c *TapApiServiceApiOAuth2Connector) GetInvitations() ([]string, error) {
	connector := c.getApiOAuth2Connector("/users/invitations")
	result := []string{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) GetUsers() ([]userManagement.UaaUser, error) {
	connector := c.getApiOAuth2Connector("/users")
	result := []userManagement.UaaUser{}
	_, err := brokerHttp.GetModel(connector, http.StatusOK, &result)
	return result, err
}

func (c *TapApiServiceApiOAuth2Connector) DeleteInvitation(email string) error {
	connector := c.getApiOAuth2Connector("/users/invitations")
	body := userManagement.InvitationRequest{
		Email: email,
	}
	_, err := brokerHttp.DeleteModelWithBody(connector, body, http.StatusNoContent)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) DeleteUser(email string) error {
	connector := c.getApiOAuth2Connector("/users")
	body := userManagement.InvitationRequest{
		Email: email,
	}
	_, err := brokerHttp.DeleteModelWithBody(connector, body, http.StatusNoContent)
	return err
}

func (c *TapApiServiceApiOAuth2Connector) ChangeCurrentUserPassword(password, newPassword string) error {
	body := userManagement.ChangePasswordRequest{
		CurrentPasswd: password,
		NewPasswd:     newPassword,
	}
	connector := c.getApiOAuth2Connector("/users/current/password")
	_, err := brokerHttp.PutModel(connector, body, http.StatusOK, "")
	return err
}

func (c *TapApiServiceApiOAuth2Connector) ExposeService(serviceId string, exposed bool) ([]string, int, error) {
	connector := c.getApiOAuth2Connector("/services/%s/expose", serviceId)
	request := models.ExposureRequest{
		Exposed: exposed,
	}
	result := &[]string{}
	status, err := brokerHttp.PutModel(connector, request, http.StatusOK, result)
	return *result, status, err
}
