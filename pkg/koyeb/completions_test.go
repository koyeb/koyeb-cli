package koyeb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterCompletions(t *testing.T) {
	values := []string{"prod-app", "prod-web", "staging-app", "b1a2c3d4", ""}

	cases := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{"empty prefix returns everything but empty values", "", []string{"prod-app", "prod-web", "staging-app", "b1a2c3d4"}},
		{"prefix filters", "prod", []string{"prod-app", "prod-web"}},
		{"prefix matches at least one", "b1a2", []string{"b1a2c3d4"}},
		{"prefix matches nothing", "zzz", nil},
		{"full value still matches", "prod-app", []string{"prod-app"}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.expected, filterCompletions(values, c.toComplete))
		})
	}
}
