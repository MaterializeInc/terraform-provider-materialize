package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransformIdWithRegion(t *testing.T) {
	SetRegionFromHostname("localhost")
	testCases := []map[string]interface{}{
		{
			"input":    "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
			"expected": "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
		},
		{
			"input":    "GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
			"expected": "aws/us-east-1:GRANT DEFAULT|SCHEMA|u1|u1|||USAGE",
		},
		{
			"input":    "aws/us-east-1:u1",
			"expected": "aws/us-east-1:u1",
		},
		{
			"input":    "u1",
			"expected": "aws/us-east-1:u1",
		},
	}
	for tc, _ := range testCases {
		c := testCases[tc]
		o := TransformIdWithRegion(c["input"].(string))
		assert.Equal(t, o, c["expected"].(string))
	}
}
