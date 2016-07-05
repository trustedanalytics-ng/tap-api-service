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
	"fmt"
	"net/http"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

func appendMetadataToService(applicationMetadata []catalogModels.Metadata, instanceMetadata []catalogModels.Metadata, offeringRequest models.CreateOfferingFromApplicationRequest) []catalogModels.Metadata {
	serviceMetadata := append(applicationMetadata, instanceMetadata...)
	serviceMetadata = append(serviceMetadata, catalogModels.Metadata{Id: "displayName", Value: offeringRequest.OfferingDisplayName})
	serviceMetadata = append(serviceMetadata, catalogModels.Metadata{Id: "description", Value: offeringRequest.Description})
	serviceMetadata = append(serviceMetadata, catalogModels.Metadata{Id: "APPLICATION_ID", Value: offeringRequest.ApplicationId})
	return serviceMetadata
}

func prepareDefaultPlan(dependencies []catalogModels.ServiceDependency) catalogModels.ServicePlan {
	plan := catalogModels.ServicePlan{
		Name:         "standard",
		Dependencies: dependencies,
	}
	return plan
}

func bindingToDependence(binding catalogModels.InstanceBindings) (catalogModels.ServiceDependency, int, error) {
	dependence := catalogModels.ServiceDependency{}
	instance, status, err := BrokerConfig.CatalogApi.GetInstance(binding.Id)
	if err != nil {
		return dependence, status, err
	}
	if instance.Type != catalogModels.InstanceTypeService {
		return dependence, http.StatusBadRequest, fmt.Errorf("This application has dependencies type: %v currently supported types: %v", instance.Type, catalogModels.InstanceTypeService)
	}

	dependentService, status, err := BrokerConfig.CatalogApi.GetService(instance.ClassId)
	if err != nil {
		return dependence, status, err
	}

	dependentPlanId := catalogModels.GetValueFromMetadata(instance.Metadata, catalogModels.OFFERING_PLAN_ID)
	dependentPlan, status, err := BrokerConfig.CatalogApi.GetServicePlan(dependentService.Id, dependentPlanId)
	if err != nil {
		return dependence, status, err
	}
	dependence.ServiceId = dependentService.Id
	dependence.ServiceName = dependentService.Name
	dependence.PlanId = dependentPlan.Id
	dependence.PlanName = dependentPlan.Name

	return dependence, http.StatusOK, nil
}
