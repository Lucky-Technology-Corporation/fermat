package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dummySecrets string = `
{
    "test": {
        "DB": "hi"
    },
    "prod": {
        "DB":  "there"
    }
}
`

func Test_ReadSecrets(t *testing.T) {
	secrets, err := ReadSecrets(strings.NewReader(dummySecrets))
	assert.Nil(t, err)
	assert.Equal(t, secrets.Test["DB"], "hi")
}

func Test_WriteThenReadSecrets(t *testing.T) {
	var serialized strings.Builder
	secrets := &Secrets{
		Test: map[string]string{"DB": "one"},
		Prod: map[string]string{"DB": "two"},
	}
	secrets.SaveSecrets(&serialized)
	deserialized, _ := ReadSecrets(strings.NewReader(serialized.String()))

	assert.Equal(t, secrets, deserialized)
}
