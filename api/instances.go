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
	"strings"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	apiApplicationInstanceMemoryDefault     = "256MB"
	apiApplicationInstanceDiskQuotaDefault  = "1024MB"
	waitingForImageApplicationInstanceState = "WAITING_FOR_IMAGE"
	errorFetchImageState                    = "ERROR_FETCH"
	errorFetchImageType                     = "-"
)

func (c *Context) AddCatalogInstanceFromApiServiceInstance(apiServiceInstance models.ServiceInstanceRequest) (catalogModels.Instance, int, error) {
	catalogInstance := catalogModels.Instance{}
	catalogInstance.Name = apiServiceInstance.Name
	catalogInstance.Bindings = apiServiceInstance.Bindings
	catalogInstance.Metadata = apiServiceInstance.Metadata
	catalogInstance.Type = apiServiceInstance.Type
	catalogInstance.ClassId = apiServiceInstance.OfferingId
	catalogInstance.State = catalogModels.InstanceStateRequested
	catalogInstance.AuditTrail = c.getAuditTrail()

	serviceId := apiServiceInstance.OfferingId

	if catalogInstance.Type == catalogModels.InstanceTypeService {
		service, status, err := BrokerConfig.CatalogApi.GetService(serviceId)
		if err != nil {
			logger.Errorf("GetService with Id %s from catalog error: %v", serviceId, err)
			return catalogInstance, status, err
		}

		applicationImage := catalogModels.GetValueFromMetadata(service.Metadata, catalogModels.APPLICATION_IMAGE_ADDRESS)
		logger.Info("application image: " + applicationImage)
		if len(applicationImage) > 0 {
			catalogInstance.Metadata = append(catalogInstance.Metadata, catalogModels.Metadata{Id: catalogModels.APPLICATION_IMAGE_ADDRESS, Value: applicationImage})
		}

		return BrokerConfig.CatalogApi.AddServiceInstance(serviceId, catalogInstance)
	} else {
		return BrokerConfig.CatalogApi.AddServiceBrokerInstance(serviceId, catalogInstance)
	}
}

func UpdateInstanceStateInCatalog(instanceId, username string, stateToSet, currentState catalogModels.InstanceState) (catalogModels.Instance, int, error) {
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata("", currentState, stateToSet)
	if err != nil {
		return catalogModels.Instance{}, 400, err
	}
	patches[0].Username = username
	return BrokerConfig.CatalogApi.UpdateInstance(instanceId, patches)
}

func (c *Context) createServiceBrokerInstance(offerings []catalogModels.Service, brokerName, templateId string) (catalogModels.Instance, int, error) {
	instance := models.ServiceInstanceRequest{}
	instance.Type = catalogModels.InstanceTypeServiceBroker
	instance.OfferingId = brokerServiceOfferingId
	instance.Name = brokerName
	instance.Metadata = append(instance.Metadata, catalogModels.Metadata{Id: catalogModels.BROKER_TEMPLATE_ID, Value: templateId})

	for _, offering := range offerings {
		metadata := catalogModels.Metadata{
			Id:    catalogModels.GetPrefixedOfferingName(offering.Name),
			Value: offering.Id,
		}
		instance.Metadata = append(instance.Metadata, metadata)
	}
	return c.AddCatalogInstanceFromApiServiceInstance(instance)
}

func getServiceBrokeInstanceMetadata(serviceBrokerInstanceId string) []catalogModels.Metadata {
	return []catalogModels.Metadata{
		{Id: templateModels.PlaceholderBrokerShortInstanceID, Value: commonHttp.UuidToShortDnsName(serviceBrokerInstanceId)},
		{Id: templateModels.PlaceholderBrokerInstanceID, Value: serviceBrokerInstanceId},
	}
}

func getPlanByInstanceMetadata(service catalogModels.Service, metadata []catalogModels.Metadata, instanceName string) (catalogModels.ServicePlan, error) {
	planId := catalogModels.GetValueFromMetadata(metadata, catalogModels.OFFERING_PLAN_ID)
	for _, plan := range service.Plans {
		if plan.Id == planId {
			return plan, nil
		}
	}
	return catalogModels.ServicePlan{}, fmt.Errorf("plan with id %q does not exist in offering %q with id %q - instance: %q", planId, service.Name, service.Id, instanceName)
}

// ConvertToApiServiceInstance merges information from catalog's service and instance models to apiServiceInstance model
func ConvertToApiServiceInstance(service catalogModels.Service, instance catalogModels.Instance) (models.ServiceInstance, error) {
	apiServiceInstance := models.ServiceInstance{}
	if instance.Type == catalogModels.InstanceTypeService {
		planOfInstance, err := getPlanByInstanceMetadata(service, instance.Metadata, instance.Name)
		if err != nil {
			return apiServiceInstance, err
		}
		apiServiceInstance.ServicePlanName = planOfInstance.Name
	}

	apiServiceInstance.Id = instance.Id
	apiServiceInstance.Name = instance.Name
	apiServiceInstance.Type = instance.Type
	apiServiceInstance.ServiceName = service.Name
	apiServiceInstance.Bindings = instance.Bindings
	apiServiceInstance.Metadata = instance.Metadata
	apiServiceInstance.AuditTrail = instance.AuditTrail
	apiServiceInstance.State = instance.State
	apiServiceInstance.OfferingId = service.Id
	return apiServiceInstance, nil
}

func ParseToApiServiceInstances(services []catalogModels.Service, instances []catalogModels.Instance) []models.ServiceInstance {
	apiServiceInstances := []models.ServiceInstance{}
	for _, instance := range instances {
		service, err := getServiceConnectedWithInstance(services, instance)
		if err != nil {
			logger.Error(err)
			continue
		}

		apiServiceInstance, err := ConvertToApiServiceInstance(service, instance)
		if err != nil {
			logger.Warning(err)
			continue
		}

		apiServiceInstances = append(apiServiceInstances, apiServiceInstance)
	}
	return apiServiceInstances
}

func ConvertToApiServiceAppInstance(application catalogModels.Application, applicationInstance catalogModels.Instance, image catalogModels.Image) (apiServiceAppInstance models.ApplicationInstance) {
	apiServiceAppInstance.Id = application.Id
	apiServiceAppInstance.Name = application.Name
	apiServiceAppInstance.Type = applicationInstance.Type
	apiServiceAppInstance.Bindings = applicationInstance.Bindings
	apiServiceAppInstance.Metadata = applicationInstance.Metadata
	apiServiceAppInstance.AuditTrail = applicationInstance.AuditTrail
	apiServiceAppInstance.State = getAppInstanceState(applicationInstance.State, image.State)

	apiServiceAppInstance.Memory = apiApplicationInstanceMemoryDefault
	apiServiceAppInstance.DiskQuota = apiApplicationInstanceDiskQuotaDefault
	apiServiceAppInstance.RunningInstances = application.Replication

	hosts := catalogModels.GetValueFromMetadata(applicationInstance.Metadata, "urls")
	apiServiceAppInstance.Urls = strings.Split(hosts, ",")

	apiServiceAppInstance.Replication = application.Replication

	apiServiceAppInstance.ImageType = image.Type
	apiServiceAppInstance.ImageState = image.State
	return apiServiceAppInstance
}

func getAppInstanceState(instanceState catalogModels.InstanceState, imageState catalogModels.ImageState) catalogModels.InstanceState {
	if instanceState == "" {
		if imageState == catalogModels.ImageStateError || imageState == errorFetchImageState {
			return catalogModels.InstanceStateFailure
		} else {
			return waitingForImageApplicationInstanceState
		}
	} else {
		return instanceState
	}
}

func ParseToApiApplicationInstances(applications []catalogModels.Application, instances []catalogModels.Instance) ([]models.ApplicationInstance, error) {
	apiApplicationInstances := []models.ApplicationInstance{}

	for _, app := range applications {

		var image catalogModels.Image
		if app.ImageId == "" {
			logger.Warning("Found null reference to Image in Application: " + app.Name)
			image.State = errorFetchImageState
			image.Type = errorFetchImageType
		} else {
			var err error
			image, _, err = BrokerConfig.CatalogApi.GetImage(app.ImageId)
			if err != nil {
				errorMessage := fmt.Sprintf("Cannot fetch image %s from Catalog: %s", app.ImageId, err.Error())
				err = errors.New(errorMessage)
				return apiApplicationInstances, err
			}
		}

		apiApplicationInstance := ConvertToApiServiceAppInstance(app, getInstanceConnectedWithApplication(app, instances), image)
		apiApplicationInstances = append(apiApplicationInstances, apiApplicationInstance)
	}

	return apiApplicationInstances, nil
}

func getServiceConnectedWithInstance(services []catalogModels.Service, instance catalogModels.Instance) (catalogModels.Service, error) {
	for _, service := range services {
		if instance.ClassId == service.Id {
			return service, nil
		}
	}
	return catalogModels.Service{}, errors.New("classId: " + instance.ClassId + " of instance: " + instance.Id + " cannot be found in services!")
}

func getInstanceConnectedWithApplication(application catalogModels.Application, instances []catalogModels.Instance) catalogModels.Instance {
	for _, instance := range instances {
		if instance.ClassId == application.Id {
			return instance
		}
	}
	return catalogModels.Instance{}
}

func limitInstanceNumber(instances int) error {
	if instances > 5 {
		return errors.New("Maximum allowed replication is 5")
	}
	if instances < 0 {
		return errors.New("Minimum allowed replication is 0")
	}
	return nil
}
