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
	"fmt"
	"net/http"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const AcceptedRequest = "Accepted"

func RestartInstance(instanceId, username string, rw web.ResponseWriter, req *web.Request) {
	instance, status, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	message := fmt.Sprintf("RestartInstance request made by: %s", username)
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata(message, instance.State, catalogModels.InstanceStateReconfiguration)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	_, status, err = BrokerConfig.CatalogApi.UpdateInstance(instanceId, patches)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, containerBrokerModels.MessageResponse{Message: AcceptedRequest}, http.StatusAccepted)
}

func StartInstance(instanceId, username string, rw web.ResponseWriter, req *web.Request) {
	message := fmt.Sprintf("StartInstance request made by: %s", username)
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata(message, catalogModels.InstanceStateStopped, catalogModels.InstanceStateStartReq)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	_, status, err := BrokerConfig.CatalogApi.UpdateInstance(instanceId, patches)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, containerBrokerModels.MessageResponse{Message: AcceptedRequest}, http.StatusAccepted)
}

func StopInstance(instanceId, username string, rw web.ResponseWriter, req *web.Request) {
	if status, err := stopInstance(instanceId, username); err != nil {
		commonHttp.GenericRespond(status, rw, err)
	}
	commonHttp.WriteJson(rw, containerBrokerModels.MessageResponse{Message: AcceptedRequest}, http.StatusAccepted)
}

func stopInstance(instanceId, username string) (int, error) {
	message := fmt.Sprintf("StopInstance request made by: %s", username)
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata(message, catalogModels.InstanceStateRunning, catalogModels.InstanceStateStopReq)
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, status, err := BrokerConfig.CatalogApi.UpdateInstance(instanceId, patches)
	if err != nil {
		return status, err
	}

	return http.StatusAccepted, nil
}

func ScaleInstance(instanceId, username string, rw web.ResponseWriter, req *web.Request) {
	scaleReq := models.ScaleApplicationRequest{}

	if err := ReadJsonAndValidate(req, &scaleReq); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	if err := limitInstanceNumber(scaleReq.Replicas); err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	instance, status, err := BrokerConfig.CatalogApi.GetInstance(instanceId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	patch, err := builder.MakePatch("Replication", scaleReq.Replicas, catalogModels.OperationUpdate)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	_, status, err = BrokerConfig.CatalogApi.UpdateApplication(instance.ClassId, []catalogModels.Patch{patch})
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	message := fmt.Sprintf("ScaleInstance request made by: %s", username)
	patches, err := builder.MakePatchesForInstanceStateAndLastStateMetadata(message, instance.State, catalogModels.InstanceStateReconfiguration)
	if err != nil {
		commonHttp.Respond400(rw, err)
		return
	}

	_, status, err = BrokerConfig.CatalogApi.UpdateInstance(instanceId, patches)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, containerBrokerModels.MessageResponse{Message: AcceptedRequest}, http.StatusAccepted)
}
