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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFilterServiceInstancesByName(t *testing.T) {
	Convey("Test empty array return empty array", t, func() {
		result := FilterServiceInstancesByName([]ServiceInstance{}, "test")
		So(result, ShouldResemble, []ServiceInstance{})
	})

	Convey("Test one matching by lowercase name instance returned from array", t, func() {
		instances := []ServiceInstance{{ServiceName: "TEST", Id: "1"}}
		result := FilterServiceInstancesByName(instances, "test")

		So(len(result), ShouldEqual, 1)
		So(result[0].Id, ShouldEqual, "1")
	})

	Convey("Test many matching by lowercase and upercase instances returned from array", t, func() {
		instances := []ServiceInstance{
			{ServiceName: "TEST", Id: "1"},
			{ServiceName: "TeST", Id: "2"},
			{ServiceName: "test", Id: "3"},
			{ServiceName: "notest", Id: "4"},
		}
		result := FilterServiceInstancesByName(instances, "test")

		So(len(result), ShouldEqual, 3)
		So(result[2].Id, ShouldEqual, "3")
	})
}

func TestFilterServiceInstancesByPlanName(t *testing.T) {
	Convey("Test empty array return empty array", t, func() {
		result := FilterServiceInstancesByPlanName([]ServiceInstance{}, "test")
		So(result, ShouldResemble, []ServiceInstance{})
	})

	Convey("Test one matching by lowercase name instance returned from array", t, func() {
		instances := []ServiceInstance{{ServiceName: "Test", ServicePlanName: "TEST"}}
		result := FilterServiceInstancesByPlanName(instances, "test")

		So(len(result), ShouldEqual, 1)
		So(result[0].ServiceName, ShouldEqual, "Test")
	})

	Convey("Test many matching by lowercase and upercase instances returned from array", t, func() {
		instances := []ServiceInstance{
			{ServiceName: "TEST", Id: "1", ServicePlanName: "test"},
			{ServiceName: "TeST", Id: "2", ServicePlanName: "TEST"},
			{ServiceName: "test", Id: "3", ServicePlanName: "Test"},
			{ServiceName: "notest", Id: "4", ServicePlanName: "test2"},
		}
		result := FilterServiceInstancesByPlanName(instances, "test")

		So(len(result), ShouldEqual, 3)
		So(result[2].Id, ShouldEqual, "3")
	})
}
