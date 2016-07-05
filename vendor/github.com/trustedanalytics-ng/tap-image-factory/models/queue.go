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

import "github.com/trustedanalytics-ng/tap-catalog/models"

const (
	IMAGE_FACTORY_QUEUE_NAME        = "tap-image-factory"
	IMAGE_FACTORY_IMAGE_ROUTING_KEY = "image"
)

type ImageGetResponse struct {
	ImageId    string `json:"id"`
	Type       string `json:"type"`
	State      string `json:"state"`
	AuditTrail models.AuditTrail
}

type ImageStatePutRequest struct {
	State string `json:"state"`
}

type BuildImagePostRequest struct {
	ImageId string `json:"id"`
}
