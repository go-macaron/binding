// Copyright 2014 martini-contrib/binding Authors
// Copyright 2014 Unknwon
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package binding

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Unknwon/macaron"
	. "github.com/smartystreets/goconvey/convey"
)

var validationTestCases = []validationTestCase{
	{
		description: "No errors",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "ID required",
		data: BlogPost{
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"id"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "Embedded struct field required",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Content: "Content given, but title is required",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: RequiredError,
				Message:        "Required",
			},
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
	{
		description: "Nested struct field required",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"name"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "Required field missing in nested struct pointer",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
			Coauthor: &Person{},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"name"},
				Classification: RequiredError,
				Message:        "Required",
			},
		},
	},
	{
		description: "All required fields specified in nested struct pointer",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Behold The Title!",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
			Coauthor: &Person{
				Name: "Jeremy Saenz",
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "Custom validation should put an error",
		data: BlogPost{
			Id: 1,
			Post: Post{
				Title:   "Too short",
				Content: "And some content",
			},
			Author: Person{
				Name: "Matt Holt",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
	{
		description: "List Validation",
		data: []BlogPost{
			BlogPost{
				Id: 1,
				Post: Post{
					Title:   "First Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
			BlogPost{
				Id: 2,
				Post: Post{
					Title:   "Second Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
		},
		expectedErrors: Errors{},
	},
	{
		description: "List Validation w/ Errors",
		data: []BlogPost{
			BlogPost{
				Id: 1,
				Post: Post{
					Title:   "First Post",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
			BlogPost{
				Id: 2,
				Post: Post{
					Title:   "Too Short",
					Content: "And some content",
				},
				Author: Person{
					Name: "Leeor Aharon",
				},
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"title"},
				Classification: "LengthError",
				Message:        "Life is too short",
			},
		},
	},
	{
		description: "List of invalid custom validations",
		data: []SadForm{
			SadForm{
				AlphaDash:    ",",
				AlphaDashDot: ",",
				MinSize:      ",",
				MinSizeSlice: []string{",", ","},
				MaxSize:      ",,",
				MaxSizeSlice: []string{",", ","},
				Email:        ",",
				Url:          ",",
				UrlEmpty:     "",
			},
		},
		expectedErrors: Errors{
			Error{
				FieldNames:     []string{"AlphaDash"},
				Classification: "AlphaDashError",
				Message:        "AlphaDash",
			},
			Error{
				FieldNames:     []string{"AlphaDashDot"},
				Classification: "AlphaDashDot",
				Message:        "AlphaDashDot",
			},
			Error{
				FieldNames:     []string{"MinSize"},
				Classification: "MinSize",
				Message:        "MinSize",
			},
			Error{
				FieldNames:     []string{"MinSize"},
				Classification: "MinSize",
				Message:        "MinSize",
			},
			Error{
				FieldNames:     []string{"MaxSize"},
				Classification: "MaxSize",
				Message:        "MaxSize",
			},
			Error{
				FieldNames:     []string{"MaxSize"},
				Classification: "MaxSize",
				Message:        "MaxSize",
			},
			Error{
				FieldNames:     []string{"Email"},
				Classification: "Email",
				Message:        "Email",
			},
			Error{
				FieldNames:     []string{"Url"},
				Classification: "Url",
				Message:        "Url",
			},
		},
	},
	{
		description: "List of valid custom validations",
		data: []SadForm{
			SadForm{
				AlphaDash:    "123-456",
				AlphaDashDot: "123.456",
				MinSize:      "12345",
				MinSizeSlice: []string{"1", "2", "3", "4", "5"},
				MaxSize:      "1",
				MaxSizeSlice: []string{"1"},
				Email:        "123@456.com",
				Url:          "http://123.456",
			},
		},
	},
}

func Test_Validation(t *testing.T) {
	Convey("Test validation", t, func() {
		for _, testCase := range validationTestCases {
			performValidationTest(t, testCase)
		}
	})
}

func performValidationTest(t *testing.T, testCase validationTestCase) {
	httpRecorder := httptest.NewRecorder()
	m := macaron.Classic()

	m.Post(testRoute, Validate(testCase.data), func(actual Errors) {
		So(fmt.Sprintf("%+v", actual), ShouldEqual, fmt.Sprintf("%+v", testCase.expectedErrors))
	})

	req, err := http.NewRequest("POST", testRoute, nil)
	if err != nil {
		panic(err)
	}

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check methods and paths")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

type (
	validationTestCase struct {
		description    string
		data           interface{}
		expectedErrors Errors
	}
)
