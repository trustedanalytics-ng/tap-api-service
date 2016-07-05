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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gocraft/web"

	"github.com/trustedanalytics-ng/tap-api-service/models"
	"github.com/trustedanalytics-ng/tap-catalog/builder"
	catalogModels "github.com/trustedanalytics-ng/tap-catalog/models"
	containerBrokerModels "github.com/trustedanalytics-ng/tap-container-broker/models"
	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
	imageFactoryModels "github.com/trustedanalytics-ng/tap-image-factory/models"
	templateRepositoryModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

func (c *Context) CreateApplicationInstance(rw web.ResponseWriter, req *web.Request) {

	// PARSE DATA
	logger.Debug("started")
	err := req.ParseMultipartForm(defaultMaxMemory)
	if err != nil {
		logger.Error("ParseMultipartForm failed!")
		commonHttp.Respond400(rw, err)
		return
	}

	manifest, err := c.getValidatedParsedManifest(req)
	if err != nil {
		logger.Error("getValidatedParsedManifest failed: ", err)
		commonHttp.Respond400(rw, getValidatedParsedManifestReadableError(err))
		return
	}

	if len(coreComponentsContainerVersions) == 0 {
		var status int
		coreComponentsContainerVersions, status, err = BrokerConfig.ContainerBrokerApi.GetVersions()
		if err != nil {
			logger.Infof("Unable to get platform versions %s", err)
			commonHttp.GenericRespond(status, rw, err)
			return
		}
	}

	if isNameUsedInCoreComponents(manifest.Name, coreComponentsContainerVersions) {
		commonHttp.Respond400(rw, fmt.Errorf("Cannot create application with name: %s, "+
			"this name is already used by core platform component", manifest.Name))
		return
	}

	// VALIDATE DEPENDENCY
	instanceDependencies, err := getInstanceDependencyForBindings(manifest.Bindings)
	if err != nil {
		logger.Error("getInstanceDependencyForBindings failed!")
		commonHttp.Respond400(rw, err)
		return
	}

	// READ FILE TAR.GZ
	blob, handler, err := req.FormFile("blob")
	if err != nil {
		errorMessage := "failed to read blob: " + err.Error()
		logger.Error(errorMessage)
		commonHttp.Respond500(rw, errors.New(errorMessage))
		return
	}
	defer blob.Close()
	logger.Infof("Read %v", handler.Filename)

	// ADD APP TO CATALOG
	application := c.getApplicationEntity(manifest, instanceDependencies)

	responseApplication, status, err := BrokerConfig.CatalogApi.AddApplication(application)
	if err != nil {
		logger.Error("CatalogApi.AddApplication failed")
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	// ADD IMAGE TO CATALOG
	imageId := catalogModels.GenerateImageId(responseApplication.Id)

	image := catalogModels.Image{
		Id:         imageId,
		Type:       manifest.ImageType,
		BlobType:   catalogModels.BlobTypeTarGz,
		AuditTrail: c.getAuditTrail(),
	}

	if _, status, err = BrokerConfig.CatalogApi.AddImage(image); err != nil {
		logger.Error("CatalogApi.AddImage failed")
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	if responseApplication, status, err = UpdateApplicationImageId(responseApplication.Id, imageId, c.Username); err != nil {
		logger.Errorf("UpdateApplicationImageId failed - Application.Id: %s , Image.Id: %s", responseApplication.Id, imageId)
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	// SAVE FILE

	if err = BrokerConfig.BlobStoreApi.StoreBlob(imageId, blob); err != nil {
		logger.Error("BlobStoreApi.StoreBlob failed")
		commonHttp.Respond500(rw, err)
		return
	}

	_, err = updateImageState(imageId, catalogModels.ImageStatePending, c.Username)
	if err != nil {
		logger.Errorf("updateImageState failed - Image.Id: %s", imageId)
		commonHttp.GenericRespond(http.StatusInternalServerError, rw, err)
		return
	}

	commonHttp.WriteJson(rw, responseApplication, http.StatusAccepted)
}

func isNameUsedInCoreComponents(name string, coreComponents []containerBrokerModels.VersionsResponse) bool {
	for _, component := range coreComponents {
		if name == component.Name {
			return true
		}
	}
	return false
}

func (c *Context) getApplicationEntity(manifest *models.Manifest, instanceDependencies []catalogModels.InstanceDependency) catalogModels.Application {
	return catalogModels.Application{
		Name:                 manifest.Name,
		Replication:          manifest.Instances,
		TemplateId:           genericApplicationTemplateID,
		InstanceDependencies: instanceDependencies,
		Metadata:             manifest.Metadata,
		AuditTrail:           c.getAuditTrail(),
	}
}

func (c *Context) parseTemplate(imageId string, req *web.Request) (*templateRepositoryModels.Template, error) {
	templateContent, err := c.readFormFile(req, "template")
	if err != nil {
		return nil, err
	}
	templateContent = bytes.Replace(templateContent, []byte("$gen_image_id"), []byte(imageId), -1)
	template := templateRepositoryModels.Template{}
	err = json.Unmarshal(templateContent, &template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (c *Context) getValidatedParsedManifest(req *web.Request) (*models.Manifest, error) {
	manifest, err := c.parseManifest(req)
	if err != nil {
		logger.Debug("parseManifest failed!")
		return &models.Manifest{}, err
	}

	err = c.validateManifest(manifest)
	if err != nil {
		logger.Debug("validateManifest failed!")
		return &models.Manifest{}, err
	}
	return manifest, nil
}

func (c *Context) parseManifest(req *web.Request) (*models.Manifest, error) {
	imageContent, err := c.readFormFile(req, "manifest")
	if err != nil {
		logger.Debug("readFormFile failed!")
		return nil, err
	}
	manifest := models.Manifest{}
	err = json.Unmarshal(imageContent, &manifest)
	if err != nil {
		logger.Debug("Unmarshal failed!")
		return nil, err
	}
	logger.Debugf("read manifest: %+v\n", manifest)
	return &manifest, nil
}

func (c *Context) parseOffering(rw web.ResponseWriter, req *web.Request) (catalogModels.Service, error) {
	offeringContent, err := c.readFormFile(req, "offering")
	if err != nil {
		commonHttp.Respond500(rw, err)
		return catalogModels.Service{}, err
	}
	offering := catalogModels.Service{}
	err = json.Unmarshal(offeringContent, &offering)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return catalogModels.Service{}, err
	}

	return offering, nil
}

func (c *Context) validateManifest(manifest *models.Manifest) error {
	if err := validateImageType(manifest.ImageType); err != nil {
		return err
	}
	if err := validateInstancesNumber(manifest.Instances); err != nil {
		return err
	}

	return nil
}

func (c *Context) parseFileTypeFromFileHandler(fileHandler *multipart.FileHeader) (catalogModels.BlobType, error) {
	extractedFileType := strings.ToUpper(filepath.Ext(fileHandler.Filename))
	extractedFileType = strings.Replace(extractedFileType, ".", "", -1)
	if extractedFileType == string(catalogModels.BlobTypeJar) {
		return catalogModels.BlobTypeJar, nil
	}

	return catalogModels.BlobType(""), errors.New(fmt.Sprintf("unrecognizable binary format %s in file name %s ", extractedFileType, fileHandler.Filename))
}

func validateImageType(imageType catalogModels.ImageType) error {
	allowedImageTypes := []catalogModels.ImageType{}
	for key := range imageFactoryModels.ImagesMap {
		allowedImageTypes = append(allowedImageTypes, key)
	}
	allowed := false
	for _, allowedType := range allowedImageTypes {
		if imageType == allowedType {
			allowed = true
			continue
		}
	}
	if !allowed {
		return fmt.Errorf("Type %v not allowed. Allowed types include: %v", imageType, allowedImageTypes)
	}
	return nil
}

func validateInstancesNumber(instancesNumber int) error {
	return limitInstanceNumber(instancesNumber)
}

func getInstanceDependencyForBindings(bindings []string) ([]catalogModels.InstanceDependency, error) {
	res := make([]catalogModels.InstanceDependency, len(bindings))

	servicesInstances, _, err := BrokerConfig.CatalogApi.ListServicesInstances()
	if err != nil {
		err = fmt.Errorf("cannot fetch application instances from catalog: %v", err)
		return nil, err
	}

	notFound := []string{}
	for i, binding := range bindings {
		found := false
		for _, instance := range servicesInstances {
			if instance.Name == binding {
				res[i].Id = instance.Id
				found = true
				break
			}
		}
		if !found {
			notFound = append(notFound, binding)
		}
	}

	if len(notFound) != 0 {
		return nil, errors.New("following service instances don't exist: " + strings.Join(notFound, ", "))
	}
	return res, nil
}

func (c *Context) readFormFile(req *web.Request, name string) ([]byte, error) {
	file, handler, err := req.FormFile(name)
	if err != nil {
		return []byte("{}"), err
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return content, err
	}
	logger.Infof("Read %v", handler.Filename)
	return content, nil
}

func (c *Context) GetApplicationInstances(rw web.ResponseWriter, req *web.Request) {
	appFilter := commonHttp.CreateItemFilter(req)
	apiApplicationInstances, err := getApplicationInstances(appFilter)
	if err != nil {
		commonHttp.Respond500(rw, err)
		return
	}

	commonHttp.WriteJson(rw, apiApplicationInstances, http.StatusOK)
}

func (c *Context) GetApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	apiApplicationInstance, err := getApplicationInstance(applicationId)
	if err != nil {
		commonHttp.GenericRespond(getStatusError(err), rw, err)
		return
	}

	commonHttp.WriteJson(rw, apiApplicationInstance, http.StatusOK)
}

func getApplicationInstances(filter *commonHttp.ItemFilter) ([]models.ApplicationInstance, error) {
	applicationInstances, _, err := BrokerConfig.CatalogApi.ListApplicationsInstances()
	if err != nil {
		err = errors.New("Cannot fetch application instances from Catalog: %s" + err.Error())
		return nil, err
	}

	applications, _, err := BrokerConfig.CatalogApi.ListApplications(filter)
	if err != nil {
		err = errors.New("Cannot fetch applications from Catalog: " + err.Error())
		return nil, err
	}

	apiApplicationInstances, err := ParseToApiApplicationInstances(applications, applicationInstances)
	if err != nil {
		err = errors.New("Cannot parse ApiApplicationInstance list: " + err.Error())
		return nil, err
	}

	return apiApplicationInstances, nil
}

func getApplicationInstance(applicationId string) (apiServiceApp models.ApplicationInstance, err error) {
	instances, _, err := BrokerConfig.CatalogApi.ListApplicationInstances(applicationId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch application %s instances from Catalog: %s", applicationId, err.Error())
		err = errors.New(errorMessage)
		return
	}

	application, _, err := BrokerConfig.CatalogApi.GetApplication(applicationId)
	if err != nil {
		errorMessage := fmt.Sprintf("Cannot fetch application %s from Catalog: %s", applicationId, err.Error())
		err = errors.New(errorMessage)
		return
	}

	apiServiceApps, err := ParseToApiApplicationInstances([]catalogModels.Application{application}, instances)
	if err != nil {
		err = errors.New("Cannot create ApiApplicationInstance: " + err.Error())
		return
	}

	if len(apiServiceApps) > 0 {
		apiServiceApp = apiServiceApps[0]
	}
	return
}

func (c *Context) DeleteApplication(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	application, status, err := BrokerConfig.CatalogApi.GetApplication(applicationId)
	if err != nil {
		commonHttp.WriteJson(rw, "", http.StatusNotFound)
		return
	}

	instances, status, err := BrokerConfig.CatalogApi.ListApplicationInstances(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}

	if len(instances) == 0 {
		if status, err = c.removeApplicationWithImageFromCatalog(application); err != nil {
			commonHttp.GenericRespond(status, rw, err)
			return
		}
		commonHttp.WriteJson(rw, "", http.StatusNoContent)
		return
	}
	instance := instances[0]

	if _, status, err = UpdateInstanceStateInCatalog(instance.Id, c.Username, catalogModels.InstanceStateDestroyReq, instance.State); err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	commonHttp.WriteJson(rw, "", http.StatusNoContent)
}

func (c *Context) removeApplicationWithImageFromCatalog(application catalogModels.Application) (int, error) {
	if status, err := BrokerConfig.CatalogApi.DeleteApplication(application.Id); err != nil {
		return status, fmt.Errorf("cannot delete application %q from Catalog: %v", application.Id, err)
	}

	if status, err := BrokerConfig.CatalogApi.DeleteImage(application.ImageId); err != nil {
		logger.Warningf("Cannot delete image %q from Catalog. Status: %d. Error: %v", application.ImageId, status, err)
	}

	return http.StatusNoContent, nil
}

func UpdateApplicationImageId(applicationId, imageId, username string) (catalogModels.Application, int, error) {
	patch, err := builder.MakePatch("ImageId", imageId, catalogModels.OperationUpdate)
	if err != nil {
		return catalogModels.Application{}, 400, err
	}
	patch.Username = username

	return BrokerConfig.CatalogApi.UpdateApplication(applicationId, []catalogModels.Patch{patch})
}

func updateImageState(imageId string, newState catalogModels.ImageState, username string) (updatedImage catalogModels.Image, err error) {
	patch, err := builder.MakePatch("State", newState, catalogModels.OperationUpdate)
	if err != nil {
		return
	}
	patch.Username = username

	updatedImage, _, err = BrokerConfig.CatalogApi.UpdateImage(imageId, []catalogModels.Patch{patch})
	return
}

func (c *Context) ScaleApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	c.makeApplicationOperation(rw, req, ScaleInstance)
}

func (c *Context) StopApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	c.makeApplicationOperation(rw, req, StopInstance)
}

func (c *Context) StartApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	applicationId := req.PathParams["applicationId"]

	application, _, err := BrokerConfig.CatalogApi.GetApplication(applicationId)
	if err != nil {
		commonHttp.Respond404(rw, err)
		return
	}

	if application.Replication == 0 {
		commonHttp.Respond400(rw, errors.New("Can't start app with 0 replicas, please scale it first"))
		return
	}

	c.makeApplicationOperation(rw, req, StartInstance)
}

func (c *Context) RestartApplicationInstance(rw web.ResponseWriter, req *web.Request) {
	c.makeApplicationOperation(rw, req, RestartInstance)
}

type applicationOperation func(instanceId, username string, rw web.ResponseWriter, req *web.Request)

func (c *Context) makeApplicationOperation(rw web.ResponseWriter, req *web.Request, operation applicationOperation) {
	applicationId := req.PathParams["applicationId"]
	validatePathParam(rw, applicationId)

	instanceId, status, err := c.getApplicationInstanceID(applicationId)
	if err != nil {
		commonHttp.GenericRespond(status, rw, err)
		return
	}
	operation(instanceId, c.Username, rw, req)
}
