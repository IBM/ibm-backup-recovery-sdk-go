/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package middlelevel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetInstanceId is a table-driven test covering the safe slice-indexing
// fix in GetInstanceId. Prior to the fix, inputs with fewer than 3 colon-
// separated segments caused an index-out-of-range panic.
func TestGetInstanceId(t *testing.T) {
	tests := []struct {
		name     string
		crn      string
		expected string
	}{
		{
			name:     "empty string returns empty without panic",
			crn:      "",
			expected: "",
		},
		{
			name:     "single segment (no colons) returns empty without panic",
			crn:      "onlyone",
			expected: "",
		},
		{
			name:     "two segments returns empty without panic",
			crn:      "a:b",
			expected: "",
		},
		{
			name:     "exactly three segments returns first segment",
			crn:      "a:b:c",
			expected: "a",
		},
		{
			name:     "valid IBM CRN returns correct instance-id segment",
			crn:      "crn:v1:bluemix:public:backup-recovery:us-south:a/account123:instance-abc::",
			expected: "instance-abc",
		},
		{
			name:     "CRN with trailing empty segments still parses correctly",
			crn:      "crn:v1:bluemix:public:backup-recovery:us-south:a/acc:inst-xyz::extra",
			expected: "inst-xyz",
		},
	}

	m := &BRSManager{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				got := m.GetInstanceId(tt.crn)
				assert.Equal(t, tt.expected, got)
			})
		})
	}
}
