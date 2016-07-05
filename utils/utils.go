/**
 * Copyright (c) 2017 Intel Corporation
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

package utils

import (
	"strings"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

func FilterInstancesByName(instances []catalogModels.Instance, name string) []catalogModels.Instance {
	name = strings.ToUpper(name)
	return filterInstanceItems(instances, func(instance catalogModels.Instance) bool {
		return (name == strings.ToUpper(instance.Name))
	})
}

func FilterInstancesByClassId(instances []catalogModels.Instance, classId string) []catalogModels.Instance {
	classId = strings.ToUpper(classId)
	return filterInstanceItems(instances, func(instance catalogModels.Instance) bool {
		return (classId == strings.ToUpper(instance.ClassId))
	})
}

type filterInstanceFunc func(catalogModels.Instance) bool

func filterInstanceItems(items []catalogModels.Instance, filter filterInstanceFunc) []catalogModels.Instance {
	filtered := []catalogModels.Instance{}
	for _, item := range items {
		if filter(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
