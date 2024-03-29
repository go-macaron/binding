// Copyright 2014 Martini Authors
// Copyright 2014 The Macaron Authors
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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/macaron.v1"
	"gopkg.in/yaml.v3"
)

var yamlTestCases = []yamlTestCase{
	{
		description:         "Happy path",
		shouldSucceedOnYaml: true,
		payload: `title: Glorious Post Title
content: Lorem ipsum dolor sit amet`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Happy path with interface",
		shouldSucceedOnYaml: true,
		withInterface:       true,
		payload: `title: Glorious Post Title
content: Lorem ipsum dolor sit amet`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Nil payload",
		shouldSucceedOnYaml: false,
		payload:             `-nil-`,
		contentType:         _YAML_CONTENT_TYPE,
		expected:            Post{},
	},
	{
		description:         "Empty payload",
		shouldSucceedOnYaml: false,
		payload:             ``,
		contentType:         _YAML_CONTENT_TYPE,
		expected:            Post{},
	},
	{
		description:         "Empty content type",
		shouldSucceedOnYaml: true,
		shouldFailOnBind:    true,
		payload: `title: Glorious Post Title
content: Lorem ipsum dolor sit amet`,
		contentType: ``,
		expected:    Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Unsupported content type",
		shouldSucceedOnYaml: true,
		shouldFailOnBind:    true,
		payload: `title: Glorious Post Title
content: Lorem ipsum dolor sit amet`,
		contentType: `BoGuS`,
		expected:    Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:         "Malformed YAML",
		shouldSucceedOnYaml: false,
		payload:             `title`,
		contentType:         _YAML_CONTENT_TYPE,
		expected:            Post{},
	},
	{
		description:         "Deserialization with nested and embedded struct",
		shouldSucceedOnYaml: true,
		payload: `
post:
  title: Glorious Post Title
id: 1
author:
  name: Matt Holt
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Deserialization with nested and embedded struct with interface",
		shouldSucceedOnYaml: true,
		withInterface:       true,
		payload: `
post:
  title: Glorious Post Title
id: 1
author:
  name: Matt Holt
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Required nested struct field not specified",
		shouldSucceedOnYaml: false,
		payload: `
post:
  title: Glorious Post Title
id: 1
author:
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    BlogPost{Post: Post{Title: "Glorious Post Title"}, Id: 1},
	},
	{
		description:         "Required embedded struct field not specified",
		shouldSucceedOnYaml: false,
		payload: `
id: 1
author:
  name: Matt Holt
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    BlogPost{Id: 1, Author: Person{Name: "Matt Holt"}},
	},
	{
		description:         "Slice of Posts",
		shouldSucceedOnYaml: true,
		payload: `
- title: First Post
- title: Second Post
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    []Post{Post{Title: "First Post"}, Post{Title: "Second Post"}},
	},
	{
		description:         "Slice of structs",
		shouldSucceedOnYaml: true,
		payload: `
name: group1
people:
  - name: awoods
  - name: anthony
`,
		contentType: _YAML_CONTENT_TYPE,
		expected:    Group{Name: "group1", People: []Person{Person{Name: "awoods"}, Person{Name: "anthony"}}},
	},
}

func Test_Yaml(t *testing.T) {
	Convey("Test YAML", t, func() {
		for _, testCase := range yamlTestCases {
			fmt.Println(testCase.description)
			data, _ := yaml.Marshal(testCase.expected)
			fmt.Println(string(data))
			performYamlTest(t, Yaml, testCase)
		}
	})
}

func performYamlTest(t *testing.T, binder handlerFunc, testCase yamlTestCase) {
	var payload io.Reader
	httpRecorder := httptest.NewRecorder()
	m := macaron.Classic()

	yamlTestHandler := func(actual interface{}, errs Errors) {
		if testCase.shouldSucceedOnYaml && len(errs) > 0 {
			So(len(errs), ShouldEqual, 0)
		} else if !testCase.shouldSucceedOnYaml && len(errs) == 0 {
			So(len(errs), ShouldNotEqual, 0)
		}
		So(fmt.Sprintf("%+v", actual), ShouldEqual, fmt.Sprintf("%+v", testCase.expected))
	}

	switch testCase.expected.(type) {
	case []Post:
		if testCase.withInterface {
			m.Post(testRoute, binder([]Post{}, (*modeler)(nil)), func(actual []Post, iface modeler, errs Errors) {

				for _, a := range actual {
					So(a.Title, ShouldEqual, iface.Model())
					yamlTestHandler(a, errs)
				}
			})
		} else {
			m.Post(testRoute, binder([]Post{}), func(actual []Post, errs Errors) {
				yamlTestHandler(actual, errs)
			})
		}

	case Post:
		if testCase.withInterface {
			m.Post(testRoute, binder(Post{}, (*modeler)(nil)), func(actual Post, iface modeler, errs Errors) {
				So(actual.Title, ShouldEqual, iface.Model())
				yamlTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(Post{}), func(actual Post, errs Errors) {
				yamlTestHandler(actual, errs)
			})
		}

	case BlogPost:
		if testCase.withInterface {
			m.Post(testRoute, binder(BlogPost{}, (*modeler)(nil)), func(actual BlogPost, iface modeler, errs Errors) {
				So(actual.Title, ShouldEqual, iface.Model())
				yamlTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(BlogPost{}), func(actual BlogPost, errs Errors) {
				yamlTestHandler(actual, errs)
			})
		}
	case Group:
		if testCase.withInterface {
			m.Post(testRoute, binder(Group{}, (*modeler)(nil)), func(actual Group, iface modeler, errs Errors) {
				So(actual.Name, ShouldEqual, iface.Model())
				yamlTestHandler(actual, errs)
			})
		} else {
			m.Post(testRoute, binder(Group{}), func(actual Group, errs Errors) {
				yamlTestHandler(actual, errs)
			})
		}
	}

	if testCase.payload == "-nil-" {
		payload = nil
	} else {
		payload = strings.NewReader(testCase.payload)
	}

	req, err := http.NewRequest("POST", testRoute, payload)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check method and path")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	default:
		if testCase.shouldSucceedOnYaml &&
			httpRecorder.Code != http.StatusOK &&
			!testCase.shouldFailOnBind {
			So(httpRecorder.Code, ShouldEqual, http.StatusOK)
		}
	}
}

type (
	yamlTestCase struct {
		description         string
		withInterface       bool
		shouldSucceedOnYaml bool
		shouldFailOnBind    bool
		payload             string
		contentType         string
		expected            interface{}
	}
)
