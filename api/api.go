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
	"os"
	"strings"
	"time"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	uaaApi "github.com/trustedanalytics-ng/tap-api-service/uaa-connector"
	userManagementApi "github.com/trustedanalytics-ng/tap-api-service/user-management-connector"
	"github.com/trustedanalytics-ng/tap-api-service/utils"
	blobStoreApi "github.com/trustedanalytics-ng/tap-blob-store/client"
	catalogApi "github.com/trustedanalytics-ng/tap-catalog/client"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	"github.com/trustedanalytics-ng/tap-container-broker/client"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	commonLogger "github.com/trustedanalytics-ng/tap-go-common/logger"
	"github.com/trustedanalytics-ng/tap-go-common/util"
	imageFactoryApi "github.com/trustedanalytics-ng/tap-image-factory/client"
	templateRepositoryApi "github.com/trustedanalytics-ng/tap-template-repository/client"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	genericServiceTemplateIDDefault     = "0e7d1029-216c-454d-8caa-33a91c5b02df"
	genericApplicationTemplateIDDefault = "52b31177-fef1-49af-7aa4-90382f7be43e"
)

var (
	errNoInstanceOfApplication      = "there are no instances of application with id %s yet."
	errTooManyInstanceOfApplication = "there are too many instances of application with id %s yet."
	errNoServiceInstanceInCatalog   = "cannot find service instance with id %s in Catalog: %s"
	errInstanceIsNotAService        = "specified instance id %s is not pointing to a service instance"
)

var genericServiceTemplateID = util.GetEnvValueOrDefault("GENERIC_SERVICE_TEMPLATE_ID", genericServiceTemplateIDDefault)
var genericApplicationTemplateID = util.GetEnvValueOrDefault("GENERIC_APPLICATION_TEMPLATE_ID", genericApplicationTemplateIDDefault)

type Config struct {
	TemplateRepositoryApi    templateRepositoryApi.TemplateRepository
	BlobStoreApi             blobStoreApi.TapBlobStoreApi
	CatalogApi               catalogApi.TapCatalogApi
	ContainerBrokerApi       api.TapContainerBrokerApi
	ImageFactoryApi          imageFactoryApi.TapApiImageFactoryApi
	UaaApi                   uaaApi.UaaApi
	UserManagementApiFactory userManagementApi.UserManagementFactory
}
type Context struct {
	*models.Context
	Domain           string
	TapVersion       string
	CliVersion       string
	CdhVersion       string
	K8sVersion       string
	CoreOrganization string
	Username         string
}

// broker service instance doesn't have an offer
const brokerServiceOfferingId = ""
const mockInstanceIdForNotCreatedYetInstance = "api-service-fake-instance-id"

var BrokerConfig *Config
var logger, _ = commonLogger.InitLogger("api")

func route(router *web.Router, context *Context, oauthMiddlewareActivated bool) {
	routeUser(router, context, oauthMiddlewareActivated)
	routeAdmin(router, context, oauthMiddlewareActivated)
}

func routeUser(router *web.Router, context *Context, oauthMiddlewareActivated bool) {
	apiRouter := router.Subrouter(*context, "")
	if oauthMiddlewareActivated {
		apiRouter.Middleware(context.Oauth2AuthorizeMiddleware)
	}

	apiRouter.Get("/platform_info", context.GetPlatformInfo)
	apiRouter.Get("/platform_components", context.GetPlatformComponents)

	apiRouter.Get("/offerings", context.GetCatalog)
	apiRouter.Post("/offerings/application", context.CreateOfferingFromApplication)
	apiRouter.Get("/offerings/:offeringId", context.GetCatalogItem)

	apiRouter.Get("/applications", context.GetApplicationInstances)
	apiRouter.Post("/applications", context.CreateApplicationInstance)
	apiRouter.Get("/applications/:applicationId", context.GetApplicationInstance)
	apiRouter.Delete("/applications/:applicationId", context.DeleteApplication)
	apiRouter.Get("/applications/:applicationId/logs", context.GetApplicationInstanceLogs)
	apiRouter.Put("/applications/:applicationId/scale", context.ScaleApplicationInstance)
	apiRouter.Put("/applications/:applicationId/stop", context.StopApplicationInstance)
	apiRouter.Put("/applications/:applicationId/start", context.StartApplicationInstance)
	apiRouter.Put("/applications/:applicationId/restart", context.RestartApplicationInstance)
	apiRouter.Get("/applications/:applicationId/bindings", context.GetApplicationInstanceBindings)
	apiRouter.Post("/applications/:applicationId/bindings", context.BindToApplicationInstance)
	apiRouter.Delete("/applications/:applicationId/bindings/services/:serviceId", context.UnbindServiceFromApplicationInstance)
	apiRouter.Delete("/applications/:dstApplicationId/bindings/applications/:srcApplicationId", context.UnbindApplicationFromApplicationInstance)

	apiRouter.Get("/services", context.GetServicesInstances)
	apiRouter.Post("/services", context.CreateServiceInstance)
	apiRouter.Get("/services/:serviceId", context.GetServiceInstance)
	apiRouter.Delete("/services/:serviceId", context.DeleteInstance)
	apiRouter.Get("/services/:serviceId/logs", context.GetServiceInstanceLogs)
	apiRouter.Put("/services/:serviceId/stop", context.StopServiceInstance)
	apiRouter.Put("/services/:serviceId/start", context.StartServiceInstance)
	apiRouter.Put("/services/:serviceId/restart", context.RestartServiceInstance)
	apiRouter.Get("/services/:serviceId/bindings", context.GetServiceInstanceBindings)
	apiRouter.Post("/services/:serviceId/bindings", context.BindToServiceInstance)
	apiRouter.Delete("/services/:dstServiceId/bindings/services/:srcServiceId", context.UnbindServiceFromServiceInstance)
	apiRouter.Delete("/services/:serviceId/bindings/applications/:applicationId", context.UnbindApplicationFromServiceInstance)
	apiRouter.Get("/services/:serviceId/credentials", context.GetServiceInstanceCredentials)
	apiRouter.Put("/services/:instanceId/expose", context.Expose)

	apiRouter.Post("/users/invitations/resend", context.ResendUserInvitation)
	apiRouter.Get("/users/invitations", context.ListInvitations)
	apiRouter.Delete("/users/invitations", context.DeleteUserInvitation)
	apiRouter.Get("/users", context.ListUsers)
	apiRouter.Delete("/users", context.DeleteUser)
	apiRouter.Put("/users/current/password", context.ChangeCurrentUserPassword)

	apiRouter.Get("/resources/cli/:resourceId", context.GetTapCliResource)
}

func routeAdmin(router *web.Router, context *Context, oauthMiddlewareActivated bool) {
	adminRouter := router.Subrouter(*context, "")
	if oauthMiddlewareActivated {
		adminRouter.Middleware(context.Oauth2AuthorizeAdminMiddleware)
	}

	adminRouter.Post("/offerings/binary", context.CreateOfferingFromBinary)
	adminRouter.Post("/offerings", context.CreateOffering)
	adminRouter.Delete("/offerings/:offeringId", context.DeleteOffering)

	adminRouter.Post("/users/invitations", context.InviteUser)
}

func (c *Context) Introduce(rw web.ResponseWriter, req *web.Request) {
	rw.Header().Add("X-Platform", "TAP")
	rw.WriteHeader(http.StatusOK)
}

func (c *Context) CheckBrokerConfig(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	if BrokerConfig == nil {
		commonHttp.Respond500(rw, errors.New("BrokerConfig not set!"))
	}
	next(rw, req)
}

func (c *Context) CreateOffering(rw web.ResponseWriter, req *web.Request) {
	resultOfferings := []catalogModels.Service{}

	serviceWithTemplate := models.ServiceDeploy{}
	if err := ReadJsonAndValidate(req, &serviceWithTemplate); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	if err := validateServices(serviceWithTemplate); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	template, err := models.ConvertRawTemplateToTemplate(serviceWithTemplate.Template)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}
	isServiceBrokerOffering := templateModels.IsServiceBrokerTemplate(template)

	if isServiceBrokerOffering {
		err := validateIfBrokerNameCanBeUsed(serviceWithTemplate.BrokerName)
		if err != nil {
			commonHttp.Respond400(rw, err)
			return
		}
	}

	responseTemplate, status, err := c.AddTemplateToCatalog()
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	status, err = AddTemplateToTemplateRepository(responseTemplate.Id, serviceWithTemplate.Template)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	_, status, err = UpdateTemplateState(responseTemplate.Id, c.Username, catalogModels.TemplateStateInProgress, catalogModels.TemplateStateReady)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	for _, offering := range serviceWithTemplate.Services {
		responseService, status, err := c.AddServiceInDeployingState(offering, responseTemplate.Id)
		if err != nil {
			commonHttp.GenericRespond(status, rw, err)
			return
		}
		logger.Debug("Response body from Catalog.AddService: ", responseService)

		if !isServiceBrokerOffering {
			responseService, status, err = UpdateServiceState(responseService.Id, c.Username, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateReady)
			if err != nil {
				commonHttp.GenericRespond(status, rw, err)
				return
			}
			logger.Debug("Response body from Catalog.UpdateService: ", responseService)
		}
		resultOfferings = append(resultOfferings, responseService)
	}

	if isServiceBrokerOffering {
		serviceBrokerInstance, status, err := c.createServiceBrokerInstance(resultOfferings, serviceWithTemplate.BrokerName, responseTemplate.Id)
		if err != nil {
			commonHttp.GenericRespond(status, rw, fmt.Errorf("cannot add service broker instance to Catalog: %v", err))
			return
		}
		logger.Debug("Service broker instance created. Response Catalog.AddInstance: ", serviceBrokerInstance)

		for _, offering := range resultOfferings {
			for _, metadata := range getServiceBrokeInstanceMetadata(serviceBrokerInstance.Id) {
				if _, status, err := UpdateServiceMetadataInCatalog(offering.Id, metadata); err != nil {
					commonHttp.GenericRespond(status, rw, err)
					return
				}
			}
		}
	}
	commonHttp.WriteJson(rw, resultOfferings, http.StatusAccepted)
}

func (c *Context) CreateOfferingFromBinary(rw web.ResponseWriter, req *web.Request) {
	logger.Info("started creating offering from binary")
	if err := req.ParseMultipartForm(defaultMaxMemory); err != nil {
		logger.Error("ParseMultipartForm failed!")
		commonHttp.Respond400(rw, err)
		return
	}

	offering, err := c.parseOffering(rw, req)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}
	err = validateSingleOffering(offering)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	manifest, err := c.parseManifest(req)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}
	err = c.validateManifest(manifest)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	blob, handler, err := req.FormFile("blob")
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}
	defer blob.Close()
	logger.Infof("Read %v", handler.Filename)

	blobType, err := c.parseFileTypeFromFileHandler(handler)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	offering.State = "DEPLOYING"
	offering.AuditTrail = c.getAuditTrail()
	offeringFromCatalog, status, err := BrokerConfig.CatalogApi.AddService(offering)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	logger.Debugf("Response body from Catalog.AddService: %v", offeringFromCatalog)

	imageId := catalogModels.ConstructImageIdForUserOffering(offeringFromCatalog.Id)
	image := catalogModels.Image{
		Id:         imageId,
		Type:       manifest.ImageType,
		BlobType:   blobType,
		AuditTrail: c.getAuditTrail(),
	}
	_, status, err = BrokerConfig.CatalogApi.AddImage(image)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	if err = BrokerConfig.BlobStoreApi.StoreBlob(imageId, blob); err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	_, err = updateImageState(imageId, catalogModels.ImageStatePending, c.Username)
	if err != nil {
		logger.Errorf("updateImageState failed - Image.Id: %s", imageId)
		commonHttp.GenericRespond(http.StatusInternalServerError, rw, err)
		return
	}

	logger.Info("creating offering from binary finished successfully")
	commonHttp.WriteJson(rw, offeringFromCatalog, http.StatusAccepted)
}

func (c *Context) DeleteOffering(rw web.ResponseWriter, req *web.Request) {
	offeringId := req.PathParams["offeringId"]

	service, template, status, err := getServiceWithTemplate(offeringId)
	if err != nil || status != http.StatusOK {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	if service.State != catalogModels.ServiceStateOffline {
		if service, status, err = UpdateServiceState(offeringId, c.Username, service.State, catalogModels.ServiceStateOffline); err != nil {
			commonHttp.GenericRespond(status, rw, err)
			return
		}
	}

	if templateModels.IsServiceBrokerTemplate(template) {
		if err := c.prepareForServiceBrokerOfferingRemoval(service); err != nil {
			commonHttp.Respond500(rw, err)
			return
		}
	}

	if status, err = deleteService(offeringId); err != nil || status != http.StatusNoContent {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	// delete template if it's possible - ie no other service uses it
	if status, err = deleteTemplate(service.TemplateId); err != nil && status != http.StatusMethodNotAllowed {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	commonHttp.WriteJson(rw, "", http.StatusAccepted)
}

func getServiceWithTemplate(serviceId string) (service catalogModels.Service, template templateModels.Template, status int, err error) {
	service, status, err = BrokerConfig.CatalogApi.GetService(serviceId)
	if err != nil {
		err = fmt.Errorf("cannot fetch service from Catalog: %v", err)
		return
	}

	templateId := service.TemplateId
	template, status, err = BrokerConfig.TemplateRepositoryApi.GenerateParsedTemplate(templateId,
		"api-service-fake-template-id", templateModels.EMPTY_PLAN_NAME, nil)
	if err != nil {
		err = fmt.Errorf("cannot fetch service templateType from Template Repository: %v", err)
		return
	}

	return
}

func (c *Context) prepareForServiceBrokerOfferingRemoval(service catalogModels.Service) error {
	brokerId := getServiceBrokerInstanceId(service)
	if brokerId == "" {
		logger.Warningf("service %q doesn't have %q key in metadata", service.Id, templateModels.PlaceholderBrokerInstanceID)
		return nil
	}

	instance, _, err := BrokerConfig.CatalogApi.GetInstance(brokerId)
	if err != nil {
		logger.Warning(err.Error())
		return nil
	}

	if err := deleteOfferingFromServiceBrokerInstance(instance, service.Name); err != nil {
		return err
	}

	instance, _, err = BrokerConfig.CatalogApi.GetInstance(brokerId)
	if err != nil {
		logger.Warning(err.Error())
		return nil
	}

	if getBrokerOfferingsCount(instance) == 0 {
		return c.deleteServiceBrokerInstance(instance, c.Username)
	}

	return nil
}

func (c *Context) deleteServiceBrokerInstance(instance catalogModels.Instance, username string) error {
	logger.Infof("Deleting service broker instance %q by %q", instance.Id, username)
	if instance.State == catalogModels.InstanceStateRunning {
		if _, err := stopInstance(instance.Id, username); err != nil {
			logger.Error(err.Error())
			return err
		}
		if err := waitForInstanceState(instance.Id, catalogModels.InstanceStateStopped); err != nil {
			return err
		}
	}
	if _, err := c.deleteInstance(instance.Id); err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
}

func waitForInstanceState(instanceId string, state catalogModels.InstanceState) error {
	waitingForInstanceStateRetries, _ := util.GetUint32EnvValueOrDefault(WaitingForInstanceStateRetries, WaitingForInstanceStateRetriesDefault)

	var i uint32 = 0
	for ; i < waitingForInstanceStateRetries; i++ {
		instance, _, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
		if err != nil {
			err = fmt.Errorf("cannot retrieve Instance %q from Catalog: %v", instanceId, err)
			logger.Error(err.Error())
			return err
		}
		if instance.State == state {
			break
		}
		time.Sleep(time.Second)
		logger.Debugf("Iteration %d. of waiting for state %q of instance %q. Current state: %q", i+2, state, instanceId, instance.State)
	}
	return nil
}

func getServiceBrokerInstanceId(service catalogModels.Service) string {
	return catalogModels.GetValueFromMetadata(service.Metadata, templateModels.PlaceholderBrokerInstanceID)
}

func getBrokerOfferingsCount(instance catalogModels.Instance) int {
	result := 0
	offeringPrefix := catalogModels.BROKER_OFFERING_PREFIX
	for _, meta := range instance.Metadata {
		if strings.HasPrefix(meta.Id, offeringPrefix) {
			result++
		}
	}

	return result
}

func deleteOfferingFromServiceBrokerInstance(instance catalogModels.Instance, serviceName string) error {
	key := catalogModels.GetPrefixedOfferingName(serviceName)
	status, err := deleteInstanceMetadataInCatalog(instance.Id, key)
	if status == http.StatusNotFound {
		logger.Warningf("service broker instance %q doesn't have offering %q entry in metadata", instance.Id, serviceName)
		return nil
	}
	return err
}

func isThereServiceInstance(serviceId string) (bool, error) {
	instances, _, err := BrokerConfig.CatalogApi.ListInstances()
	if err != nil {
		return true, errors.New("Cannot fetch instances: " + err.Error())
	}

	for _, instance := range instances {
		if instance.ClassId == serviceId {
			return true, nil
		}
	}

	return false, nil
}

func deleteService(serviceId string) (int, error) {
	serviceInstanceExist, err := isThereServiceInstance(serviceId)
	if err != nil {
		err = errors.New("Cannot determine if there is a service instance: " + err.Error())
		return http.StatusMethodNotAllowed, err
	} else if serviceInstanceExist {
		err = errors.New("There is an instance of service being deleted")
		return http.StatusMethodNotAllowed, err
	}

	return BrokerConfig.CatalogApi.DeleteService(serviceId)
}

func isTemplateUsed(templateId string) (bool, error) {
	services, _, err := BrokerConfig.CatalogApi.GetServices()
	if err != nil {
		return true, errors.New("Cannot fetch services from Catalog: " + err.Error())
	}

	for _, service := range services {
		if service.TemplateId == templateId {
			return true, nil
		}
	}

	return false, nil
}

func isGenericTemplate(templateId string) bool {
	return templateId == genericApplicationTemplateID || templateId == genericServiceTemplateID
}

func deleteTemplate(templateId string) (int, error) {
	isUsed, err := isTemplateUsed(templateId)
	if err != nil {
		err = errors.New("Couldn't determine if template is used by any service when deleting offering: " + err.Error())
		return http.StatusInternalServerError, err
	} else if isUsed || isGenericTemplate(templateId) {
		err = errors.New("Cannot delete template because it's used by some service")
		return http.StatusMethodNotAllowed, err
	}

	if status, err := BrokerConfig.TemplateRepositoryApi.DeleteTemplate(templateId); err != nil || status != http.StatusNoContent {
		err = errors.New(fmt.Sprintf("Cannot delete template %q from Template Repository: %v", templateId, err))
		return status, err
	} else {
		return status, nil
	}
}

func (c *Context) GetCatalog(rw web.ResponseWriter, req *web.Request) {
	services, _, err := BrokerConfig.CatalogApi.GetServices()
	if err != nil {
		err = errors.New("Cannot fetch services from Catalog: " + err.Error())
		commonHttp.Respond500(rw, err)
		return
	}

	result := []models.Offering{}

	for _, service := range services {
		brokerInstance, status, err := getServiceBrokerInstanceForService(service)
		if err != nil {
			commonHttp.GenericRespond(status, rw, errors.New("getServiceBrokerInstanceForService error: "+err.Error()))
			return
		}
		apiService := ParseServiceToOffering(service, brokerInstance)
		result = append(result, apiService)
	}
	commonHttp.WriteJson(rw, result, http.StatusOK)
}

func (c *Context) GetCatalogItem(rw web.ResponseWriter, req *web.Request) {
	serviceId := req.PathParams["offeringId"]

	service, status, err := BrokerConfig.CatalogApi.GetService(serviceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, errors.New("Cannot fetch service from Catalog: "+err.Error()))
		return
	}

	brokerInstance, status, err := getServiceBrokerInstanceForService(service)
	if err != nil {
		commonHttp.GenericRespond(status, rw, errors.New("getServiceBrokerInstanceForService error: "+err.Error()))
		return
	}

	apiServices := ParseServiceToOffering(service, brokerInstance)

	commonHttp.WriteJson(rw, apiServices, http.StatusOK)
}

func (c *Context) CreateServiceInstance(rw web.ResponseWriter, req *web.Request) {
	apiServiceInstance := models.ServiceInstanceRequest{}

	if err := ReadJsonAndValidate(req, &apiServiceInstance); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	service, _, err := BrokerConfig.CatalogApi.GetService(apiServiceInstance.OfferingId)
	if err != nil {
		commonHttp.Respond500(rw, err)
	}

	if service.State != catalogModels.ServiceStateReady {
		commonHttp.Respond400(rw, fmt.Errorf("service %q has inappropriate state %q", service.Id, string(service.State)))
		return
	}

	plan, err := getPlanByInstanceMetadata(service, apiServiceInstance.Metadata, apiServiceInstance.Name)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	for _, dependency := range plan.Dependencies {
		depInstance := prepareInstanceFromDependency(apiServiceInstance.Name, dependency)

		if instance, status, err := c.AddCatalogInstanceFromApiServiceInstance(depInstance); err != nil {
			err = fmt.Errorf("cannot add service instance %s dependency %s to Catalog: %v", apiServiceInstance.OfferingId, dependency.ServiceName, err)
			commonHttp.GenericRespond(status, rw, err)
			return
		} else {
			apiServiceInstance.Bindings = append(apiServiceInstance.Bindings, catalogModels.InstanceBindings{Id: instance.Id})
		}
	}

	instance, status, err := c.AddCatalogInstanceFromApiServiceInstance(apiServiceInstance)
	if err != nil {
		commonHttp.GenericRespond(status, rw, fmt.Errorf("cannot add service instance %s to Catalog: %v", apiServiceInstance.OfferingId, err))
		return
	}

	response, err := ConvertToApiServiceInstance(service, instance)
	if err != nil {
		commonHttp.Respond500(rw, err)
	}

	commonHttp.WriteJson(rw, response, http.StatusAccepted)
}

func prepareInstanceFromDependency(parentName string, dependency catalogModels.ServiceDependency) models.ServiceInstanceRequest {
	metadata := []catalogModels.Metadata{{
		Id:    catalogModels.OFFERING_PLAN_ID,
		Value: dependency.PlanId,
	}}

	return models.ServiceInstanceRequest{
		Type:       catalogModels.InstanceTypeService,
		OfferingId: dependency.ServiceId,
		Name:       parentName + "-" + dependency.ServiceName,
		Metadata:   metadata,
	}
}

func (c *Context) DeleteInstance(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]

	if status, err := c.deleteInstance(instanceId); err != nil {
		commonHttp.GenericRespond(status, rw, err)
	}

	commonHttp.WriteJson(rw, "", http.StatusAccepted)
}

func (c *Context) deleteInstance(instanceId string) (int, error) {
	instance, status, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
	if err != nil {
		return status, err
	}

	if status, err := validateDeleteInstanceBindings(instance); err != nil {
		return status, err
	}

	_, status, err = UpdateInstanceStateInCatalog(instanceId, c.Username, catalogModels.InstanceStateDestroyReq, instance.State)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot update service instance %s state to '%v' in Catalog: %s", instanceId, catalogModels.InstanceStateDestroyReq, err.Error())
		err = errors.New(errorMessage)
		return status, err
	}
	return http.StatusAccepted, nil
}

func (c *Context) StopServiceInstance(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]
	StopInstance(instanceId, c.Username, rw, req)
}

func (c *Context) StartServiceInstance(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]
	StartInstance(instanceId, c.Username, rw, req)
}

func (c *Context) RestartServiceInstance(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]
	RestartInstance(instanceId, c.Username, rw, req)
}

func validateDeleteInstanceBindings(srcInstance catalogModels.Instance) (int, error) {
	instances, status, err := BrokerConfig.CatalogApi.ListInstances()
	if err != nil {
		return status, err
	}

	for _, instance := range instances {
		for _, binding := range instance.Bindings {
			if srcInstance.Id == binding.Id {
				dstID := instance.Id
				if instance.Type == catalogModels.InstanceTypeApplication {
					dstID = instance.ClassId
				}

				return http.StatusForbidden, fmt.Errorf("Instance: %s is bound to other instance: %s, id: %s",
					srcInstance.Name, instance.Name, dstID)
			}
		}
	}
	return http.StatusOK, nil
}

func (c *Context) GetApplicationInstanceLogs(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	instanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	logs, status, err := BrokerConfig.ContainerBrokerApi.GetInstanceLogs(instanceId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch application instance %s logs from Container Broker: %s", applicationId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, logs, status)
}

func (c *Context) GetServiceInstanceLogs(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]

	logs, status, err := BrokerConfig.ContainerBrokerApi.GetInstanceLogs(instanceId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch service instance %s logs from Container Broker: %s", instanceId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, logs, status)
}

func (c *Context) checkServiceInstanceID(serviceInstanceID string) (int, error) {
	instance, status, err := BrokerConfig.CatalogApi.GetInstance(serviceInstanceID)
	if err != nil {
		return status, fmt.Errorf(errNoServiceInstanceInCatalog, serviceInstanceID, err)
	}

	if instance.Type != catalogModels.InstanceTypeService {
		return http.StatusNotFound, fmt.Errorf(errInstanceIsNotAService, serviceInstanceID)
	}

	return http.StatusOK, nil
}

func (c *Context) getApplicationInstanceID(applicationId string) (string, int, error) {
	instances, status, err := BrokerConfig.CatalogApi.ListApplicationInstances(applicationId)
	if err != nil {
		e := fmt.Errorf("cannot fetch instance of application with id %s from Catalog: %s",
			applicationId, err.Error())
		return "", status, e
	}
	if len(instances) == 0 {
		err := fmt.Errorf(errNoInstanceOfApplication, applicationId)
		return "", http.StatusNotFound, err
	}
	if len(instances) > 1 {
		err := fmt.Errorf(errTooManyInstanceOfApplication, applicationId)
		return "", http.StatusInternalServerError, err
	}
	return instances[0].Id, http.StatusOK, nil
}

func (c *Context) GetServiceInstanceCredentials(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]

	credentials, status, err := BrokerConfig.ContainerBrokerApi.GetCredentials(instanceId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch instance %s credentials from Container Broker: %s", instanceId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, credentials, http.StatusOK)
}

func (c *Context) GetServicesInstances(rw web.ResponseWriter, req *web.Request) {
	instances, status, err := BrokerConfig.CatalogApi.ListServicesInstances()
	if err != nil {
		err = errors.New("Cannot fetch service instances from Catalog: " + err.Error())
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	services, status, err := BrokerConfig.CatalogApi.GetServices()
	if err != nil {
		err = errors.New("Cannot fetch services from Catalog: " + err.Error())
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	if id := commonHttp.GetQueryParameterCaseInsensitive(req, "offeringId"); id != "" {
		instances = utils.FilterInstancesByClassId(instances, id)
	}
	if name := commonHttp.GetQueryParameterCaseInsensitive(req, "name"); name != "" {
		instances = utils.FilterInstancesByName(instances, name)
	}

	apiServiceInstances := ParseToApiServiceInstances(services, instances)

	if name := commonHttp.GetQueryParameterCaseInsensitive(req, "offeringName"); name != "" {
		apiServiceInstances = models.FilterServiceInstancesByName(apiServiceInstances, name)
	}
	if name := commonHttp.GetQueryParameterCaseInsensitive(req, "planName"); name != "" {
		apiServiceInstances = models.FilterServiceInstancesByPlanName(apiServiceInstances, name)
	}

	commonHttp.WriteJson(rw, apiServiceInstances, http.StatusOK)
}

func (c *Context) GetServiceInstance(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["serviceId"]
	validatePathParam(rw, instanceId)

	instance, status, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch instance %s from Catalog: %s", instanceId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	service, status, err := BrokerConfig.CatalogApi.GetService(instance.ClassId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch service %s from Catalog: %s", instance.ClassId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	apiServiceInstance, err := ConvertToApiServiceInstance(service, instance)
	if err != nil {
		err = errors.New("Cannot create ApiServiceInstance: " + err.Error())
		commonHttp.Respond500(rw, err)
		return
	}

	commonHttp.WriteJson(rw, apiServiceInstance, http.StatusOK)
}

func (c *Context) Expose(rw web.ResponseWriter, req *web.Request) {
	instanceId := req.PathParams["instanceId"]

	request := models.ExposureRequest{}
	if err := commonHttp.ReadJson(req, &request); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	instance, status, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, errors.New(fmt.Sprintf("cannot fetch instance %s from Catalog: %s", instanceId, err.Error())))
		return
	}

	if request.Exposed {
		exposeRequest := containerBrokerModels.ExposeRequest{
			Hostname: instance.Name,
		}
		if hosts, status, err := BrokerConfig.ContainerBrokerApi.ExposeInstance(instanceId, exposeRequest); err != nil {
			commonHttp.GenericRespond(status, rw, errors.New(fmt.Sprintf("cannot expose instance %s: %s", instanceId, err.Error())))
		} else {
			commonHttp.WriteJson(rw, hosts, http.StatusOK)
		}
	} else if request.Exposed == false {
		if status, err := BrokerConfig.ContainerBrokerApi.UnexposeInstance(instanceId); err != nil {
			commonHttp.GenericRespond(status, rw, errors.New(fmt.Sprintf("cannot unexpose instance %s: %s", instanceId, err.Error())))
			return
		}
		commonHttp.WriteJson(rw, []string{}, http.StatusOK)
	} else {
		commonHttp.GenericRespond(http.StatusBadRequest, rw, errors.New(fmt.Sprintf("exposed should be true or false")))
		return
	}
}

func (c *Context) GetTapCliResource(rw web.ResponseWriter, req *web.Request) {
	resourceId := req.PathParams["resourceId"]

	switch resourceId {
	case "linux32":
		c.serveResource(rw, req, "tap-linux32")
	case "linux64":
		c.serveResource(rw, req, "tap-linux64")
	case "windows32":
		c.serveResource(rw, req, "tap-windows32.exe")
	case "macosx64":
		c.serveResource(rw, req, "tap-macosx64.osx")
	default:
		commonHttp.GenericRespond(http.StatusNotFound, rw, fmt.Errorf("Unknown resource id: %s", resourceId))
	}
}

func (c *Context) serveResource(rw web.ResponseWriter, req *web.Request, resource string) {
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Debug("Problem with getting the current dir, using ./ instead")
		currentDir = "."
	}

	resourcePath := currentDir + "/resources/" + resource

	if _, err := os.Stat(resourcePath); err != nil {
		if os.IsNotExist(err) {
			logger.Errorf("Unable to find resource file: %s", resourcePath)
			commonHttp.GenericRespond(http.StatusInternalServerError, rw,
				fmt.Errorf("Resource is not available: %s (unable to access resource file)", resource))
		} else {
			logger.Errorf("Resource (path: %s) is not available: %s", resourcePath, err)
			commonHttp.GenericRespond(http.StatusInternalServerError, rw,
				fmt.Errorf("Resource is not available: %s (internal error)", resource))
		}
	}

	rw.Header().Set("Content-Type", "application/octet-stream")
	rw.Header().Set("Content-Disposition", "attachment;filename=\""+resource+"\"")
	http.ServeFile(rw, req.Request, resourcePath)
}

func (c *Context) CreateOfferingFromApplication(rw web.ResponseWriter, req *web.Request) {
	offeringRequest := models.CreateOfferingFromApplicationRequest{}
	err := commonHttp.ReadJson(req, &offeringRequest)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	applicationInstanceId, status, err := c.getApplicationInstanceID(offeringRequest.ApplicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	application, status, err := BrokerConfig.CatalogApi.GetApplication(offeringRequest.ApplicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	applicationInstance, status, err := BrokerConfig.CatalogApi.GetInstance(applicationInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	dependencies := []catalogModels.ServiceDependency{}
	for _, binding := range applicationInstance.Bindings {
		dependence, status, err := bindingToDependence(binding)
		if err != nil {
			commonHttp.GenericRespond(status, rw, err)
			return
		}
		dependencies = append(dependencies, dependence)
	}

	serviceMetadata := appendMetadataToService(application.Metadata, applicationInstance.Metadata, offeringRequest)
	servicePlan := prepareDefaultPlan(dependencies)

	service := catalogModels.Service{
		Name:        offeringRequest.OfferingName,
		TemplateId:  genericServiceTemplateID,
		Metadata:    serviceMetadata,
		Description: offeringRequest.Description,
		Plans:       []catalogModels.ServicePlan{servicePlan},
		Tags:        offeringRequest.Tags,
	}

	service, status, err = BrokerConfig.CatalogApi.AddService(service)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	service, status, err = UpdateServiceState(service.Id, c.Username, catalogModels.ServiceStateDeploying, catalogModels.ServiceStateReady)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, service, http.StatusOK)
}
