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
package utils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
)

func TestFilterInstancesByName(t *testing.T) {
	Convey("Test empty array return empty array", t, func() {
		result := FilterInstancesByName([]catalogModels.Instance{}, "test")
		So(result, ShouldResemble, []catalogModels.Instance{})
	})

	Convey("Test one matching by lowercase name instance returned from array", t, func() {
		instances := []catalogModels.Instance{{Name: "TEST", Id: "1"}}
		result := FilterInstancesByName(instances, "test")

		So(len(result), ShouldEqual, 1)
		So(result[0].Id, ShouldEqual, "1")
	})

	Convey("Test many matching by lowercase and upercase instances returned from array", t, func() {
		instances := []catalogModels.Instance{
			{Name: "TEST", Id: "1"},
			{Name: "TeST", Id: "2"},
			{Name: "test", Id: "3"},
			{Name: "notest", Id: "4"},
		}
		result := FilterInstancesByName(instances, "test")

		So(len(result), ShouldEqual, 3)
		So(result[2].Id, ShouldEqual, "3")
	})
}

func TestFilterInstancesByClassId(t *testing.T) {
	Convey("Test empty array return empty array", t, func() {
		result := FilterInstancesByClassId([]catalogModels.Instance{}, "test")
		So(result, ShouldResemble, []catalogModels.Instance{})
	})

	Convey("Test one matching by lowercase name instance returned from array", t, func() {
		instances := []catalogModels.Instance{{Name: "TEST", Id: "1", ClassId: "test1"}}
		result := FilterInstancesByClassId(instances, "test1")

		So(len(result), ShouldEqual, 1)
		So(result[0].Id, ShouldEqual, "1")
	})

	Convey("Test many matching by lowercase and upercase instances returned from array", t, func() {
		instances := []catalogModels.Instance{
			{Name: "TEST", Id: "1", ClassId: "test1"},
			{Name: "TeST", Id: "2", ClassId: "TEST1"},
			{Name: "test", Id: "3", ClassId: "Test1"},
			{Name: "notest", Id: "4", ClassId: "test2"},
		}
		result := FilterInstancesByClassId(instances, "test1")

		So(len(result), ShouldEqual, 3)
		So(result[2].Id, ShouldEqual, "3")
	})
}
