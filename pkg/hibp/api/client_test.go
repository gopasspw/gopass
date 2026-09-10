package api

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Example() { //nolint:testableexamples
	matches, err := Lookup("sha1sum of secret")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Number of matches: %d", matches)
}

func TestLookup(t *testing.T) { //nolint:paralleltest
	match := "match"
	noMatch := "no match"
	matchSum := sha1sum(match)
	var matchCount uint64 = 324567

	reqCnt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCnt++
		if reqCnt < 2 {
			http.Error(w, "fake error", http.StatusInternalServerError)

			return
		}
		if strings.TrimPrefix(r.URL.String(), "/range/") == matchSum[:5] {
			fmt.Fprintf(w, "%s", matchSum[5:10]+":1\r\n")         // invalid
			fmt.Fprintf(w, "%s", matchSum[5:39]+":3234879\r\n")   // invalid
			fmt.Fprintf(w, "%s", matchSum[5:]+":\r\n")            // invalid
			fmt.Fprintf(w, "%s", matchSum[5:]+"\r\n")             // invalid
			fmt.Fprintf(w, "%s:%d\r\n", matchSum[5:], matchCount) // valid

			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()
	URL = ts.URL

	// test with one entry
	count, err := Lookup(matchSum)
	require.NoError(t, err)
	assert.Equal(t, matchCount, count)

	// add another one
	count, err = Lookup(sha1sum(noMatch))
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)

	// invalid input
	count, err = Lookup("")
	require.Error(t, err)
	assert.Equal(t, uint64(0), count)
}

func TestLookupCR(t *testing.T) { //nolint:paralleltest
	match := "match"
	noMatch := "no match"
	matchSum := sha1sum(match)
	var matchCount uint64 = 324567

	reqCnt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCnt++
		if reqCnt < 2 {
			http.Error(w, "fake error", http.StatusInternalServerError)

			return
		}
		if strings.TrimPrefix(r.URL.String(), "/range/") == matchSum[:5] {
			fmt.Fprintf(w, "%s", matchSum[5:10]+":1\n")         // invalid
			fmt.Fprintf(w, "%s", matchSum[5:39]+":3234879\n")   // invalid
			fmt.Fprintf(w, "%s", matchSum[5:]+":\n")            // invalid
			fmt.Fprintf(w, "%s", matchSum[5:]+"\n")             // invalid
			fmt.Fprintf(w, "%s:%d\n", matchSum[5:], matchCount) // valid

			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer ts.Close()
	URL = ts.URL

	// test with one entry
	count, err := Lookup(matchSum)
	require.NoError(t, err)
	assert.Equal(t, matchCount, count)

	// add another one
	count, err = Lookup(sha1sum(noMatch))
	require.NoError(t, err)
	assert.Equal(t, uint64(0), count)
}

func sha1sum(data string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(data))

	return fmt.Sprintf("%X", h.Sum(nil))
}
