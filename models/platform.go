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

type PlatformInfo struct {
	ApiEndpoint      string        `json:"api_endpoint"`
	CliVersion       string        `json:"cli_version"`
	CliUrl           string        `json:"cli_url"`
	PlatformVersion  string        `json:"platform_version"`
	CoreOrganization string        `json:"core_organization"`
	ExternalTools    ExternalTools `json:"external_tools"`
	CdhVersion       string        `json:"cdh_version"`
	K8sVersion       string        `json:"k8s_version"`
}

type ExternalTools struct {
	Visualizations []PlatformTool `json:"visualizations"`
}

type PlatformComponents struct {
	Name          string `json:"name"`
	ImageVersions string `json:"imageVersion"`
	Signature     string `json:"signature"`
	AppVersion    string `json:"app_version"`
}

type PlatformTool struct {
	Name      string `json:"name"`
	Url       string `json:"url"`
	Available bool   `json:"available"`
}
