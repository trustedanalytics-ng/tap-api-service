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
package user_management_connector

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewUserManagementConnectorFactory(t *testing.T) {
	Convey("Test DeleteOffering should return nil error", t, func() {
		client, err := NewUserManagementConnectorFactory("adress")
		So(err, ShouldBeNil)
		So(client.Address, ShouldEqual, "adress")
	})
}
