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
	"reflect"
	"strings"

	"github.com/gocraft/web"
	"gopkg.in/validator.v2"

	commonHttp "github.com/trustedanalytics-ng/tap-go-common/http"
)

const keyNotFoundMessage = "Key not found"

func getStatusError(err error) (httpStatus int) {
	httpStatus = http.StatusInternalServerError
	if strings.Contains(strings.ToUpper(err.Error()), strings.ToUpper(keyNotFoundMessage)) {
		httpStatus = http.StatusNotFound
	}
	return
}
func validatePathParam(rw web.ResponseWriter, param string) {
	if param == "" {
		commonHttp.Respond400(rw, errors.New("Path param can not be empty!"))
		return
	}
}

func oneOf(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return validator.ErrUnsupported
	}

	possibleValues := strings.Split(param, ";")

	for _, value := range possibleValues {
		if st.String() == value {
			return nil
		}
	}
	return fmt.Errorf("value cannot be \"%v\" , must be one of: [%v]", st.String(), param)
}

func ReadJsonAndValidate(req *web.Request, retstruct interface{}) error {
	err := validator.SetValidationFunc("oneOf", oneOf)
	if err != nil {
		logger.Critical("failed to set 'oneOf' validation func for validator!")
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("Error reading request body:", err)
		return err
	}
	err = json.Unmarshal(body, &retstruct)
	if err != nil {
		logger.Error("Error parsing request body json:", err)
		return err
	}
	logger.Debug("Request JSON parsed as: ", retstruct)

	if err := validator.Validate(retstruct); err != nil {
		logger.Error("Error during validation of request body:", err)
		return err
	}
	return nil
}
