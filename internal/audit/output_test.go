package audit

import (
	"bytes"
	"fmt"
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
	today := time.Now().Format("2006-01-02")
	assert.Equal(t, fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
  <head>
  <meta charset="utf-8">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>gopass audit report generated on %s</title>
  <style>
#findings {
  font-family: Arial, Helvetica, sans-serif;
  border-collapse: collapse;
  width: 100%%;
}
#findings td, #findings th {
  border: 1px solid #ddd;
  padding: 8px;
}
#findings tr:nth-child(even){
  background-color: #f3f3f3;
}
#findings tr:hover {
  background-color: #ddd;
}
#findings th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: #03995D;
  color: white;
}
  </style>
</head>
<body>

Audited 1 secrets in 0s on %s.<br />

<table id="findings">
  <thead>
  <th>Secret</th>

<th>duplicate</th>

<th>hibp-api</th>

  </thead>
  <tr>
    <td>foo</td>
    <td class="warning">
        <div title="found duplicates">found duplicates</div>
    </td>
    <td class="warning">
        <div title="found match on HIBP">found match on HIBP</div>
    </td>
  </tr>
</table>
</body>
</html>
`, today, today), out.String())
}
