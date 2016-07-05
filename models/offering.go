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
	"encoding/json"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

type Offering struct {
	Name           string                   `json:"name"`
	DisplayName    string                   `json:"displayName"`
	Provider       string                   `json:"provider"`
	Url            string                   `json:"url"`
	Description    string                   `json:"description"`
	Version        string                   `json:"version"`
	Bindable       bool                     `json:"bindable"`
	Id             string                   `json:"id"`
	Tags           []string                 `json:"tags"`
	State          string                   `json:"state"`
	OfferingPlans  []OfferingPlan           `json:"offeringPlans"`
	Metadata       []catalogModels.Metadata `json:"metadata"`
	BrokerInstance ServiceInstance          `json:"broker_instance"`
}

type OfferingPlan struct {
	Name        string `json:"name"`
	Free        bool   `json:"free"`
	Description string `json:"description"`
	OfferingId  string `json:"offeringId"`
	Id          string `json:"id"`
	Active      bool   `json:"active"`
}

type simpleKubernetesBody struct {
	Type templateModels.ComponentType `json:"componentType"`
}

func ConvertRawTemplateToTemplate(rawTemplate templateModels.RawTemplate) (templateModels.Template, error) {
	result := templateModels.Template{}

	for key, val := range rawTemplate {
		if key == "body" {
			body, err := json.Marshal(val)
			if err != nil {
				return result, err
			}

			components := []simpleKubernetesBody{}
			if err = json.Unmarshal(body, &components); err != nil {
				return result, err
			}

			for _, component := range components {
				result.Body = append(result.Body, templateModels.KubernetesComponent{Type: component.Type})
			}
		}
	}
	return result, nil
}
