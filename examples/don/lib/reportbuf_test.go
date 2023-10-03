package donlib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReportBuffer_GetLatest(t *testing.T) {
	n := 10
	rb := NewReportBuffer(n)

	t.Run("empty", func(t *testing.T) {
		latest := rb.GetLatest("emptydon")
		require.Nil(t, latest)
	})

	t.Run("happy flow", func(t *testing.T) {
		testDon := "happyflowdon"
		for i := 0; i < n-1; i++ {
			require.True(t, rb.Add(testDon, MockedSignedReport{SeqNumber: int64(i + 1)}))
		}
		latest := rb.GetLatest(testDon)
		require.NotNil(t, latest)
		require.Equal(t, int64(n-1), latest.SeqNumber)
	})

	t.Run("circular flow", func(t *testing.T) {
		testDon := "circulardon"
		for i := 0; i < n; i++ {
			require.True(t, rb.Add(testDon, MockedSignedReport{SeqNumber: int64(i + 1)}))
		}
		require.True(t, rb.Add(testDon, MockedSignedReport{SeqNumber: int64(n + 1)}))
		rb.lock.Lock()
		t.Logf("buffer: %+v", rb.reports[testDon])
		rb.lock.Unlock()

		latest := rb.GetLatest(testDon)
		require.NotNil(t, latest)
		require.Equal(t, int64(n+1), latest.SeqNumber)
	})
}
