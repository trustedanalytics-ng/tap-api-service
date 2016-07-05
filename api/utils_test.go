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
	"errors"
	"fmt"
	"net/http"
	"testing"

	"gopkg.in/validator.v2"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetStatusError(t *testing.T) {
	Convey("Given some error", t, func() {
		testErr := errors.New("some error")

		Convey("When getStatusError is invoked with this error", func() {
			realStatus := getStatusError(testErr)

			Convey("getStatusError should return Internal Server Error status", func() {
				So(realStatus, ShouldEqual, http.StatusInternalServerError)
			})
		})
	})

	Convey("Given an error containing 'Key not found' message", t, func() {
		notFoundErr := errors.New("error! Key not found: /Services")

		Convey("When getStatusError is invoked with this error", func() {
			realStatus := getStatusError(notFoundErr)

			Convey("getStatusError should return Not Found status", func() {
				So(realStatus, ShouldEqual, http.StatusNotFound)
			})
		})
	})

	Convey("Given an error containing 'Key not found' message", t, func() {
		mixedCaseErr := errors.New("mixed case error! keY nOt fOUNd: /Services")

		Convey("When getStatusError is invoked with this error", func() {
			realStatus := getStatusError(mixedCaseErr)

			Convey("getStatusError should return Not Found status", func() {
				So(realStatus, ShouldEqual, http.StatusNotFound)
			})
		})
	})
}

type TestStruct struct {
	Weather string `validate:"oneOf=SUNNY;RAINY;SNOWY"`
}

type TestErrorStruct struct {
	Colour string `validate:"oneeOf=RED;YELLOW;BLUE;BLACK"`
}

func TestOneOf(t *testing.T) {
	err := validator.SetValidationFunc("oneOf", oneOf)
	if err != nil {
		t.Fatal("failed to set 'oneOf' validation func for validator!")
	}

	testCases := []struct {
		weatherPassed     string
		shouldReturnError bool
	}{
		{
			weatherPassed: "SUNNY",
		},
		{
			weatherPassed: "SNOWY",
		},
		{
			weatherPassed:     "WINDY",
			shouldReturnError: true,
		},
		{
			weatherPassed:     "rainy",
			shouldReturnError: true,
		},
	}

	for _, tc := range testCases {
		Convey(fmt.Sprintf("Given struct with weather=%v", tc.weatherPassed), t, func() {
			testData := TestStruct{
				Weather: tc.weatherPassed,
			}

			Convey("When validator is invoked", func() {
				err := validator.Validate(testData)

				if tc.shouldReturnError {
					Convey("should return error", func() {
						So(err, ShouldNotBeNil)
					})
				} else {
					Convey("should not return error", func() {
						So(err, ShouldBeNil)
					})
				}
			})
		})
	}

	Convey("Given bad defined struct", t, func() {
		testData := TestErrorStruct{
			Colour: "RED",
		}

		Convey("When validator is invoked", func() {
			err := validator.Validate(testData)

			Convey("should return error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
