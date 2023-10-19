package donlib

import (
	"testing"

	"github.com/amirylm/p2pmq/proto"
	"github.com/stretchr/testify/require"
)

func TestVerifier_validateSequence(t *testing.T) {
	t.Skip()
	n := 100
	rb := NewReportBuffer(n)

	v := NewVerifier(rb, "")

	missing := map[int]bool{
		n - 5: true,
	}
	testDon := "testdon"
	for i := 0; i < n; i++ {
		if missing[i+1] {
			continue
		}
		require.True(t, rb.Add(testDon, MockedSignedReport{SeqNumber: int64(i + 1)}))
	}

	tests := []struct {
		name string
		don  string
		seq  int64
		data []byte
		res  proto.ValidationResult
	}{
		{
			name: "empty buffer",
			don:  "emptydon",
			seq:  1,
			data: []byte("testdata"),
			res:  proto.ValidationResult_ACCEPT,
		},
		{
			name: "happy flow",
			don:  testDon,
			seq:  int64(n + 1),
			data: []byte("testdata-valid"),
			res:  proto.ValidationResult_ACCEPT,
		},
		{
			name: "under skip treshold",
			don:  testDon,
			seq:  int64(n - 5), // missing
			data: []byte("testdata-valid-almost-skipped"),
			res:  proto.ValidationResult_ACCEPT,
		},
		{
			name: "ignored message",
			don:  testDon,
			seq:  int64(n) - (skipThreshold + 2),
			data: []byte("testdata-ignored"),
			res:  proto.ValidationResult_IGNORE,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// TODO: fix
			r, err := NewMockedSignedReport(nil, tc.seq, tc.don, tc.data)
			require.NoError(t, err)
			require.Equal(t, tc.res, v.(*verifier).validateSequence(r))
		})
	}
}
