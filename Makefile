# Copyright (c) 2016 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
GOBIN=$(GOPATH)/bin
APP_DIR_LIST=$(shell go list ./... | grep -v /vendor/)
APP_NAME=tap-api-service

build: verify_gopath
	go fmt $(APP_DIR_LIST)
	CGO_ENABLED=0 go install -tags netgo .
	mkdir -p application && cp -f $(GOBIN)/$(APP_NAME) ./application/$(APP_NAME)

verify_gopath:
	@if [ -z "$(GOPATH)" ] || [ "$(GOPATH)" = "" ]; then\
		echo "GOPATH not set. You need to set GOPATH before run this command";\
		exit 1 ;\
	fi

docker_build: build_anywhere
	docker build -t tap-api-service .

push_docker: docker_build
	docker tag tap-api-service $(REPOSITORY_URL)/tap-api-service:latest
	docker push $(REPOSITORY_URL)/tap-api-service:latest

kubernetes_deploy: docker_build
	kubectl create -f configmap.yaml
	kubectl create -f service.yaml
	kubectl create -f deployment.yaml

kubernetes_update: docker_build
	kubectl delete -f deployment.yaml
	kubectl create -f deployment.yaml

deps_fetch_specific: bin/govendor
	@if [ "$(DEP_URL)" = "" ]; then\
		echo "DEP_URL not set. Run this comand as follow:";\
		echo " make deps_fetch_specific DEP_URL=github.com/nu7hatch/gouuid";\
	exit 1 ;\
	fi
	@echo "Fetching specific dependency in newest versions"
	$(GOBIN)/govendor fetch -v $(DEP_URL)

deps_update_tap: verify_gopath
	$(GOBIN)/govendor update github.com/trustedanalytics-ng/...
	$(GOBIN)/govendor remove github.com/trustedanalytics-ng/$(APP_NAME)/...
	@echo "Done"

bin/govendor: verify_gopath
	go get -v -u github.com/kardianos/govendor

test: verify_gopath
	CGO_ENABLED=0 go test -tags netgo --cover $(APP_DIR_LIST)

prepare_dirs:
	mkdir -p ./temp/src/github.com/trustedanalytics-ng/tap-api-service
	$(eval REPOFILES=$(shell pwd)/*)
	ln -sf $(REPOFILES) temp/src/github.com/trustedanalytics-ng/tap-api-service
	mkdir -p ./resources

build_anywhere: prepare_dirs
	$(eval GOPATH=$(shell cd ./temp; pwd))
	$(eval APP_DIR_LIST=$(shell GOPATH=$(GOPATH) go list ./temp/src/github.com/trustedanalytics-ng/tap-api-service/... | grep -v /vendor/))
	GOPATH=$(GOPATH) CGO_ENABLED=0 go build -tags netgo $(APP_DIR_LIST)
	rm -Rf application && mkdir application
	cp ./tap-api-service ./application/tap-api-service
	rm -Rf ./temp

install_mockgen:
	scripts/install_mockgen.sh

mock_update:
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-catalog/client/client.go -package=api -destination=api/catalog_mock_test.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-template-repository/client/client_api.go -package=api -destination=api/template_repository_mock_test.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-blob-store/client/client.go -package=api -destination=api/blob_store_mock_test.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-container-broker/client/client_api.go -package=api -destination=api/container_broker_mock_test.go
	$(GOBIN)/mockgen -source=uaa-connector/client.go -package=uaa_connector -destination=uaa-connector/uaa_mock.go
	$(GOBIN)/mockgen -source=user-management-connector/client.go -package=user_management_connector -destination=user-management-connector/user_management_mock.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-catalog/client/client.go -package=api_v2 -destination=api_v2/catalog_mock_test.go
	$(GOBIN)/mockgen -source=vendor/github.com/trustedanalytics-ng/tap-blob-store/client/client.go -package=api_v2 -destination=api_v2/blob_store_mock_test.go
	./add_license.sh
