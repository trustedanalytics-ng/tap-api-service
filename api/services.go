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
	"strings"

	apiServiceModels "github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func ParseServiceToOffering(service catalogModels.Service, brokerInstance *apiServiceModels.ServiceInstance) apiServiceModels.Offering {
	apiService := apiServiceModels.Offering{}

	apiService.Name = service.Name
	apiService.DisplayName = catalogModels.GetValueFromMetadata(service.Metadata, "displayName")
	apiService.Provider = catalogModels.GetValueFromMetadata(service.Metadata, "provider")
	apiService.Url = catalogModels.GetValueFromMetadata(service.Metadata, "url")
	apiService.Description = service.Description
	apiService.Version = catalogModels.GetValueFromMetadata(service.Metadata, "version")
	apiService.Bindable = service.Bindable
	apiService.Id = service.Id
	apiService.State = string(service.State)
	apiService.Metadata = service.Metadata
	apiService.Tags = service.Tags

	for _, plan := range service.Plans {
		servicePlan := apiServiceModels.OfferingPlan{}
		servicePlan.Name = plan.Name
		servicePlan.Description = plan.Description
		servicePlan.OfferingId = apiService.Id
		servicePlan.Id = plan.Id
		servicePlan.Active = true
		servicePlan.Free = strings.ToLower(plan.Cost) == "free"
		apiService.OfferingPlans = append(apiService.OfferingPlans, servicePlan)
	}

	if brokerInstance != nil {
		apiService.BrokerInstance = *brokerInstance
	}
	return apiService
}

func getServiceBrokerInstanceForService(service catalogModels.Service) (*apiServiceModels.ServiceInstance, int, error) {
	brokerInstanceId := catalogModels.GetValueFromMetadata(service.Metadata, templateModels.PlaceholderBrokerInstanceID)
	if brokerInstanceId != "" {
		instance, status, err := BrokerConfig.CatalogApi.GetInstance(brokerInstanceId)
		if err != nil {
			return nil, status, err
		}

		brokerInstance, err := ConvertToApiServiceInstance(service, instance)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return &brokerInstance, http.StatusOK, nil
	}
	return nil, http.StatusOK, nil
}

func (c *Context) AddServiceInDeployingState(catalogService catalogModels.Service, templateIdRelatedToThisService string) (catalogModels.Service, int, error) {
	catalogService.State = "DEPLOYING"
	catalogService.TemplateId = templateIdRelatedToThisService
	catalogService.AuditTrail = c.getAuditTrail()
	return BrokerConfig.CatalogApi.AddService(catalogService)
}

func UpdateServiceState(serviceId, username string, oldState, newState catalogModels.ServiceState) (catalogModels.Service, int, error) {
	patch, err := builder.MakePatchWithPreviousValue("State", newState, oldState, catalogModels.OperationUpdate)
	if err != nil {
		return catalogModels.Service{}, 400, err
	}
	patch.Username = username

	return BrokerConfig.CatalogApi.UpdateService(serviceId, []catalogModels.Patch{patch})
}

func UpdateServiceMetadataInCatalog(serviceId string, metadata catalogModels.Metadata) (catalogModels.Service, int, error) {
	patch, err := builder.MakePatch("Metadata", metadata, catalogModels.OperationAdd)
	if err != nil {
		return catalogModels.Service{}, 400, err
	}

	return BrokerConfig.CatalogApi.UpdateService(serviceId, []catalogModels.Patch{patch})
}

func deleteInstanceMetadataInCatalog(serviceId, key string) (int, error) {
	metadata := catalogModels.Metadata{
		Id: key, Value: "",
	}
	patch, err := builder.MakePatch("Metadata", metadata, catalogModels.OperationDelete)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	_, status, err := BrokerConfig.CatalogApi.UpdateInstance(serviceId, []catalogModels.Patch{patch})
	return status, err
}

func validateServices(serviceDeploy apiServiceModels.ServiceDeploy) error {
	if len(serviceDeploy.Services) == 0 {
		return errors.New("No service is defined")
	}
	for _, service := range serviceDeploy.Services {
		err := validateSingleOffering(service)
		if err != nil {
			return err
		}
	}
	return nil
}

func validateIfBrokerNameCanBeUsed(brokerName string) error {
	if brokerName == "" {
		return errors.New("broker name cannont be empty!")
	}

	err := catalogModels.CheckIfMatchingRegexp(brokerName, catalogModels.RegexpDnsLabelLowercase)
	if err != nil {
		return fmt.Errorf("broker name is invalid: %v", err)
	}

	instances, _, err := BrokerConfig.CatalogApi.ListInstances()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		if instance.Name == brokerName {
			return fmt.Errorf("brokerName %s is already in use!", brokerName)
		}
	}
	return nil
}

func validateSingleOffering(service catalogModels.Service) error {
	if len(service.Name) == 0 {
		return errors.New("Service name should not be empty")
	}
	if len(service.Plans) == 0 {
		return errors.New("Service must have at least one plan")
	}

	for _, plan := range service.Plans {
		for _, dependency := range plan.Dependencies {
			if dependency.ServiceId == "" || dependency.PlanId == "" {
				return errors.New(fmt.Sprintf("Dependency: %s service error: ServiceId and PlanId should not be empty!", dependency.ServiceName))
			}
			service, _, err := BrokerConfig.CatalogApi.GetService(dependency.ServiceId)
			if err != nil {
				return errors.New(fmt.Sprintf("Dependency: %s service error: %v", dependency.ServiceName, err))
			}
			if service.State != catalogModels.ServiceStateReady {
				return errors.New(fmt.Sprintf("Dependency: %s service has inappropriate state: %s", dependency.ServiceName, service.State))
			}

			planExist := false
			for _, dependencyPlan := range service.Plans {
				if dependencyPlan.Id == dependency.PlanId {
					planExist = true
					break
				}
			}

			if !planExist {
				return errors.New(fmt.Sprintf("Dependency: %s Plan does not exist!: %v", dependency.PlanName, err))
			}
		}
	}
	return nil
}
