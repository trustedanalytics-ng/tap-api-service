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
package models

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	templateModels "github.com/trustedanalytics-ng/tap-template-repository/model"
)

func TestConvertRawTemplateToTemplate(t *testing.T) {
	Convey("Test to marshal wrong type istead of array", t, func() {
		result, err := ConvertRawTemplateToTemplate(getWrongMarshaledRawTemplate())
		So(err, ShouldNotBeEmpty)
		So(result, ShouldResemble, templateModels.Template{})
	})

	Convey("Test to unmarshaling wrong json value to array", t, func() {
		result, err := ConvertRawTemplateToTemplate(getWrongUnMarshaledRawTemplate())
		So(err, ShouldNotBeEmpty)
		So(result, ShouldResemble, templateModels.Template{})
	})

	Convey("Test empty raw template return empty template", t, func() {
		result, err := ConvertRawTemplateToTemplate(templateModels.RawTemplate{})
		So(err, ShouldBeEmpty)
		So(result, ShouldResemble, templateModels.Template{})
	})

	Convey("Test raw template with 3 elements returning template with 3 element array", t, func() {
		result, err := ConvertRawTemplateToTemplate(getFakeRawTemplate())
		So(err, ShouldBeEmpty)
		So(result, ShouldResemble, getExpectedTemplate())
	})
}

func getWrongMarshaledRawTemplate() templateModels.RawTemplate {
	raw := templateModels.RawTemplate{}
	raw["body"] = func() {}
	return raw
}

func getWrongUnMarshaledRawTemplate() templateModels.RawTemplate {
	raw := templateModels.RawTemplate{}
	raw["body"] = "wrongType1"
	return raw
}

func getFakeRawTemplate() templateModels.RawTemplate {
	raw := templateModels.RawTemplate{}
	raw["body"] = [3]simpleKubernetesBody{
		{"fakeType1"},
		{"fakeType2"},
		{"fakeType3"},
	}
	return raw
}

func getExpectedTemplate() templateModels.Template {
	return templateModels.Template{
		Body: []templateModels.KubernetesComponent{
			{Type: "fakeType1"},
			{Type: "fakeType2"},
			{Type: "fakeType3"},
		},
	}
}
