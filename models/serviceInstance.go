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
package models

import (
	"strings"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

type ServiceInstance struct {
	Id              string                           `json:"id"`
	Name            string                           `json:"name"`
	Type            catalogModels.InstanceType       `json:"type"`
	OfferingId      string                           `json:"offeringId"`
	Bindings        []catalogModels.InstanceBindings `json:"bindings"`
	Metadata        []catalogModels.Metadata         `json:"metadata"`
	State           catalogModels.InstanceState      `json:"state"`
	AuditTrail      catalogModels.AuditTrail         `json:"auditTrail"`
	ServiceName     string                           `json:"serviceName"`
	ServicePlanName string                           `json:"planName"`
}

type ServiceInstanceRequest struct {
	Name       string                           `json:"name" validate:"nonzero"`
	Type       catalogModels.InstanceType       `json:"type" validate:"nonzero,oneOf=APPLICATION;SERVICE;SERVICE_BROKER"`
	OfferingId string                           `json:"offeringId" validate:"nonzero"`
	Bindings   []catalogModels.InstanceBindings `json:"bindings"`
	Metadata   []catalogModels.Metadata         `json:"metadata" validate:"nonzero"`
}

func FilterServiceInstancesByName(services []ServiceInstance, serviceName string) []ServiceInstance {
	serviceName = strings.ToUpper(serviceName)
	return filterServiceInstanceItems(services, func(service ServiceInstance) bool {
		return (serviceName == strings.ToUpper(service.ServiceName))
	})
}

func FilterServiceInstancesByPlanName(services []ServiceInstance, planName string) []ServiceInstance {
	planName = strings.ToUpper(planName)
	return filterServiceInstanceItems(services, func(service ServiceInstance) bool {
		return (planName == strings.ToUpper(service.ServicePlanName))
	})
}

type filterServiceInstanceFunc func(ServiceInstance) bool

func filterServiceInstanceItems(items []ServiceInstance, filter filterServiceInstanceFunc) []ServiceInstance {
	filtered := []ServiceInstance{}
	for _, item := range items {
		if filter(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
