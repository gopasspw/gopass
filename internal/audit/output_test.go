package audit

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTML(t *testing.T) {
	r := newReport()

	r.AddPassword("foo", "bar")
	r.SetAge("foo", time.Hour)
	r.AddFinding("foo", "duplicate", "found duplicates", "warning")
	r.AddFinding("foo", "hibp-api", "found match on HIBP", "warning")

	sr := r.Finalize()
	out := &bytes.Buffer{}
	require.NoError(t, sr.RenderHTML(out))
	assert.Equal(t, `<!DOCTYPE html>
<html lang="en">
  <head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>gopass audit report</title>
</head>
<body>
<table>
  <thead>
  <th>Secret</th>
  <th>Age</th>
  <th>Findings</th>
  </thead>
  <tr>
    <td>foo</td>
	<td>60m</td>
	<td>duplicate: found duplicates (warning)</td>
	<td>hibp-api: found match on HIBP (warning)</td>
  </tr>
</table>
</body>
</html>
`, out.String())
}
