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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

var errBadInstanceBindingRequest = errors.New("exactly one of ids has to be filled")

func (c *Context) GetApplicationInstanceBindings(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	instanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.getInstanceBindings(instanceId, rw)
}

func (c *Context) GetServiceInstanceBindings(rw web.ResponseWriter, req *web.Request) {
	serviceId := req.PathParams["serviceId"]

	status, err := c.checkServiceInstanceID(serviceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.getInstanceBindings(serviceId, rw)
}

func (c *Context) getInstanceBindings(instanceId string, rw web.ResponseWriter) {
	instances, status, err := BrokerConfig.CatalogApi.GetInstanceBindings(instanceId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch instance %s bindings from Catalog: %s", instanceId, err.Error())
		err = errors.New(errorMessage)
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	result := models.InstanceBindings{}
	for _, instance := range instances {
		var entity models.InstanceBindingsEntity
		if instance.Type == catalogModels.InstanceTypeApplication {
			entity = models.InstanceBindingsEntity{AppGUID: instance.ClassId, AppInstanceName: instance.Name}
		} else {
			entity = models.InstanceBindingsEntity{ServiceInstanceGUID: instance.Id, ServiceInstanceName: instance.Name}
		}
		resource := models.InstanceBindingsResource{InstanceBindingsEntity: entity}
		result.Resources = append(result.Resources, resource)
	}
	commonHttp.WriteJson(rw, result, status)
}

func (c *Context) BindToApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	dstInstanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.bindInstance(models.DstInstanceBind{DstInstanceId: dstInstanceId, AppGUID: applicationId}, rw, req)
}

func (c *Context) BindToServiceInstance(rw web.ResponseWriter, req *web.Request) {
	dstInstanceId := req.PathParams["serviceId"]

	status, err := c.checkServiceInstanceID(dstInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.bindInstance(models.DstInstanceBind{DstInstanceId: dstInstanceId}, rw, req)
}

func (c *Context) bindInstance(dstInstance models.DstInstanceBind, rw web.ResponseWriter, req *web.Request) {
	request := models.InstanceBindingRequest{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}
	err = json.Unmarshal(body, &request)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}
	srcInstanceId, status, err := c.getInstanceIdBasedOnOneOfIds(&request)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	response, status, err := BrokerConfig.ContainerBrokerApi.BindInstance(srcInstanceId, dstInstance.DstInstanceId)
	if err != nil {
		//message will be formated to output AppGUID if exists since container broker bind instances based on instance id
		dstInstanceId := dstInstance.DstInstanceId
		if dstInstance.AppGUID != "" {
			dstInstanceId = dstInstance.AppGUID
		}
		err = fmt.Errorf("Cannot bind instance %s to %s: %s", srcInstanceId, dstInstanceId, err.Error())
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	commonHttp.WriteJson(rw, response, http.StatusAccepted)
}

func (c *Context) getInstanceIdBasedOnOneOfIds(ids *models.InstanceBindingRequest) (string, int, error) {
	if (ids.ApplicationId != "" && ids.ServiceId != "") || (ids.ApplicationId == "" && ids.ServiceId == "") {
		return "", http.StatusBadRequest, errBadInstanceBindingRequest
	}

	if ids.ApplicationId != "" {
		return c.getApplicationInstanceID(ids.ApplicationId)
	}

	status, err := c.checkServiceInstanceID(ids.ServiceId)
	return ids.ServiceId, status, err
}

func (c *Context) UnbindServiceFromApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]
	srcInstanceId := req.PathParams["serviceId"]

	dstInstanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	status, err = c.checkServiceInstanceID(srcInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.unbindInstance(srcInstanceId, dstInstanceId, rw)
}

func (c *Context) UnbindApplicationFromApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	dstApplicationId := req.PathParams["dstApplicationId"]
	srcApplicationId := req.PathParams["srcApplicationId"]

	dstInstanceId, status, err := c.getApplicationInstanceID(dstApplicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	srcInstanceId, status, err := c.getApplicationInstanceID(srcApplicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.unbindInstance(srcInstanceId, dstInstanceId, rw)
}

func (c *Context) UnbindServiceFromServiceInstance(rw web.ResponseWriter, req *web.Request) {
	dstInstanceId := req.PathParams["dstServiceId"]
	srcInstanceId := req.PathParams["srcServiceId"]

	status, err := c.checkServiceInstanceID(dstInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	status, err = c.checkServiceInstanceID(srcInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.unbindInstance(srcInstanceId, dstInstanceId, rw)
}

func (c *Context) UnbindApplicationFromServiceInstance(rw web.ResponseWriter, req *web.Request) {
	dstInstanceId := req.PathParams["serviceId"]
	applicationId := req.PathParams["applicationId"]

	status, err := c.checkServiceInstanceID(dstInstanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	srcInstanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	c.unbindInstance(srcInstanceId, dstInstanceId, rw)
}

func (c *Context) unbindInstance(srcInstanceId, dstInstanceId string, rw web.ResponseWriter) {
	response, status, err := BrokerConfig.ContainerBrokerApi.UnbindInstance(srcInstanceId, dstInstanceId)
	if err != nil {
		err := fmt.Errorf("cannot unbind instance %q from %q: %v", dstInstanceId, srcInstanceId, err)
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, response, http.StatusAccepted)
}
