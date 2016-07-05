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
	"errors"
	"net/http"

	"github.com/gocraft/web"

	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

func (c *Context) GetApiServiceHealth(rw web.ResponseWriter, req *web.Request) {
	err := BrokerConfig.TemplateRepositoryApi.GetTemplateRepositoryHealth()
	if err != nil {
		err = errors.New("Cannot fetch Template Repository health: " + err.Error())
		commonHttp.Respond500(rw, err)
		return
	}

	_, err = BrokerConfig.CatalogApi.GetCatalogHealth()
	if err != nil {
		err = errors.New("Cannot fetch Catalog health: " + err.Error())
		commonHttp.Respond500(rw, err)
		return
	}

	_, err = BrokerConfig.ContainerBrokerApi.GetContainerBrokerHealth()
	if err != nil {
		err = errors.New("Cannot fetch Container Broker health: " + err.Error())
		commonHttp.Respond500(rw, err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}
