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
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

type Instance struct {
	Name     string                           `json:"name"`
	Type     catalogModels.InstanceType       `json:"type"`
	ClassId  string                           `json:"classId"`
	Bindings []catalogModels.InstanceBindings `json:"bindings"`
	Metadata []catalogModels.Metadata         `json:"metadata"`
}

type InstanceBindings struct {
	Resources []InstanceBindingsResource `json:"resources"`
}

type InstanceBindingsResource struct {
	InstanceBindingsEntity `json:"entity"`
}

type DstInstanceBind struct {
	AppGUID       string
	DstInstanceId string
}

type InstanceBindingRequest struct {
	ApplicationId string `json:"application_id"`
	ServiceId     string `json:"service_id"`
}

type InstanceBindingsEntity struct {
	AppGUID             string `json:"app_guid"`
	AppInstanceName     string `json:"app_instance_name"`
	ServiceInstanceGUID string `json:"service_instance_guid"`
	ServiceInstanceName string `json:"service_instance_name"`
}

type ExposureRequest struct {
	Exposed bool `json:"exposed"`
}
