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
// Automatically generated by MockGen. DO NOT EDIT!
// Source: vendor/github.com/trustedanalytics-ng/tap-template-repository/client/client_api.go

package api

import (
	gomock "github.com/golang/mock/gomock"
	model "github.com/trustedanalytics-ng/tap-template-repository/model"
)

// Mock of TemplateRepository interface
type MockTemplateRepository struct {
	ctrl     *gomock.Controller
	recorder *_MockTemplateRepositoryRecorder
}

// Recorder for MockTemplateRepository (not exported)
type _MockTemplateRepositoryRecorder struct {
	mock *MockTemplateRepository
}

func NewMockTemplateRepository(ctrl *gomock.Controller) *MockTemplateRepository {
	mock := &MockTemplateRepository{ctrl: ctrl}
	mock.recorder = &_MockTemplateRepositoryRecorder{mock}
	return mock
}

func (_m *MockTemplateRepository) EXPECT() *_MockTemplateRepositoryRecorder {
	return _m.recorder
}

func (_m *MockTemplateRepository) GenerateParsedTemplate(templateId string, uuid string, planName string, replacements map[string]string) (model.Template, int, error) {
	ret := _m.ctrl.Call(_m, "GenerateParsedTemplate", templateId, uuid, planName, replacements)
	ret0, _ := ret[0].(model.Template)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockTemplateRepositoryRecorder) GenerateParsedTemplate(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GenerateParsedTemplate", arg0, arg1, arg2, arg3)
}

func (_m *MockTemplateRepository) CreateTemplate(template model.RawTemplate) (int, error) {
	ret := _m.ctrl.Call(_m, "CreateTemplate", template)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockTemplateRepositoryRecorder) CreateTemplate(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateTemplate", arg0)
}

func (_m *MockTemplateRepository) GetRawTemplate(templateId string) (model.RawTemplate, int, error) {
	ret := _m.ctrl.Call(_m, "GetRawTemplate", templateId)
	ret0, _ := ret[0].(model.RawTemplate)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (_mr *_MockTemplateRepositoryRecorder) GetRawTemplate(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetRawTemplate", arg0)
}

func (_m *MockTemplateRepository) DeleteTemplate(templateId string) (int, error) {
	ret := _m.ctrl.Call(_m, "DeleteTemplate", templateId)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockTemplateRepositoryRecorder) DeleteTemplate(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteTemplate", arg0)
}

func (_m *MockTemplateRepository) GetTemplateRepositoryHealth() error {
	ret := _m.ctrl.Call(_m, "GetTemplateRepositoryHealth")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockTemplateRepositoryRecorder) GetTemplateRepositoryHealth() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetTemplateRepositoryHealth")
}
