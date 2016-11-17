// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/palantir/pkg/cli"
)

func TestRunErrorOutput(t *testing.T) {
	cases := []struct {
		err            error
		expectedOutput string
	}{
		// empty error should not produce any output
		{
			err:            fmt.Errorf(""),
			expectedOutput: "",
		},
		// non-empty error is printed with newline appended
		{
			err:            fmt.Errorf("foo"),
			expectedOutput: "foo\n",
		},
	}

	for i, currCase := range cases {
		app := cli.NewApp()
		app.Action = func(ctx cli.Context) error {
			return currCase.err
		}

		stderr := &bytes.Buffer{}
		app.Stderr = stderr

		exitStatus := app.Run([]string{"testApp"})
		if exitStatus == 0 {
			t.Errorf("Case %d: expected exitStatus to be non-0, was: %d", i, exitStatus)
		}

		if stderr.String() != currCase.expectedOutput {
			t.Errorf("Case %d:\nExpected: %q\nActual:   %q", i, currCase.expectedOutput, stderr.String())
		}
	}
}

func TestRunErrorHandler(t *testing.T) {
	cases := []struct {
		err              error
		handler          func(ctx cli.Context, err error) int
		expectedExitCode int
		expectedOutput   string
	}{
		// custom error handler is invoked if provided
		{
			err: fmt.Errorf(""),
			handler: func(ctx cli.Context, err error) int {
				ctx.Errorf("Error: %v\n", err)
				return 13
			},
			expectedExitCode: 13,
			expectedOutput:   "Error: \n",
		},
		// default behavior is used if custom error hander is nil
		{
			err:              fmt.Errorf("foo"),
			expectedExitCode: 1,
			expectedOutput:   "foo\n",
		},
	}

	for i, currCase := range cases {
		app := cli.NewApp()
		app.ErrorHandler = currCase.handler
		app.Action = func(ctx cli.Context) error {
			return currCase.err
		}

		stderr := &bytes.Buffer{}
		app.Stderr = stderr

		exitCode := app.Run([]string{"testApp"})
		assert.Equal(t, currCase.expectedExitCode, exitCode, "Case %d", i)
		assert.Equal(t, currCase.expectedOutput, stderr.String(), "Case %d", i)
	}
}
