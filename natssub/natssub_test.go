package natssub_test

import (
	"os"
	"testing"

	"github.com/hsmtkk/turbo-doodle/natssub"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	json, err := os.ReadFile("./sample.json")
	assert.Nil(t, err)
	want := "test/gpg4win-4.0.0.exe"
	got, err := natssub.Parse(json)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}
