package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testSol1 CLIErrorSolution = "Solution 1"
	testSol2 CLIErrorSolution = "Solution 1"
)

func TestTemplate(t *testing.T) {
	values := map[string]struct {
		err      *CLIError
		expected string
	}{
		"minimal": {
			err: &CLIError{
				What:     "Error title",
				Why:      "error message",
				Solution: testSol1,
			},
			expected: `❌ Error title: error message

🏥 How to solve the issue?
Solution 1
`,
		},
		"with_additional_info": {
			err: &CLIError{
				What:       "Error title",
				Why:        "error message",
				Additional: []string{"additional info 1", "additional info 2"},
				Solution:   testSol1,
			},
			expected: `❌ Error title: error message

🔎 Additional details
additional info 1
additional info 2

🏥 How to solve the issue?
Solution 1
`,
		},
		"with_original_error": {
			err: &CLIError{
				What:     "Error title",
				Why:      "error message",
				Orig:     fmt.Errorf("original error"),
				Solution: testSol1,
			},
			expected: `❌ Error title: error message

🏥 How to solve the issue?
Solution 1

🕦 The original error was:
original error
`,
		},
		"full": {
			err: &CLIError{
				What:       "Error title",
				Why:        "error message",
				Additional: []string{"additional info 1", "additional info 2"},
				Orig:       fmt.Errorf("original error"),
				Solution:   testSol1,
			},
			expected: `❌ Error title: error message

🔎 Additional details
additional info 1
additional info 2

🏥 How to solve the issue?
Solution 1

🕦 The original error was:
original error
`,
		},
	}

	for name, tc := range values {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}
