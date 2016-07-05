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

package api

import (
	"net/http"
	"testing"

	"github.com/gocraft/web"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/trustedanalytics-ng/tap-api-service/models"
)

func TestIntroduceRoute(t *testing.T) {

	Convey("Given api service router listening", t, func() {
		BrokerConfig = &Config{}
		r := web.New(models.Context{})
		SetupRouter(r, false)

		Convey("when root api endpoint is requested", func() {
			resp := SendGet("/api", r)

			Convey("it shall return 200 OK status", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
			})

			Convey("it shall introduce TAP by passing header", func() {
				So(resp.Header().Get("X-Platform"), ShouldEqual, "TAP")
			})
		})
	})
}
