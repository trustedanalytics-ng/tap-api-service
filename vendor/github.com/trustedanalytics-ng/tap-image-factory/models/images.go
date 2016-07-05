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

var (
	ImagesMap = map[catalogModels.ImageType]string{
		"JAVA":      "tap-base-java:java8-jessie",
		"GO":        "tap-base-binary:binary-jessie",
		"NODEJS":    "tap-base-node:node4.4-jessie",
		"PYTHON2.7": "tap-base-python:python2.7-jessie",
		"PYTHON3.4": "tap-base-python:python3.4",
	}
)
