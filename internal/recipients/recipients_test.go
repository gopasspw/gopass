package recipients

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   []string
		want string
	}{
		{
			want: "foo@bar.com\n",
			in:   []string{"foo@bar.com\n\r"},
		},
		{
			want: "baz@bar.com\nfoo@bar.com\n",
			in:   []string{"baz@bar.com", "foo@bar.com"},
		},
		{
			want: "baz@bar.com\nzab@zab.com\n",
			in:   []string{"baz@bar.com", "zab@zab.com"},
		},
	} {
		tc := tc
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			r := New()
			sort.Strings(tc.in)
			for _, k := range tc.in {
				r.Add(k)
			}
			got := string(r.Marshal())
			assert.Equal(t, tc.want, got, tc.want)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want []string
	}{
		{
			in:   "foo@bar.com",
			want: []string{"foo@bar.com"},
		},
		{
			in:   "foo@bar.com\nbaz@bar.com\n",
			want: []string{"baz@bar.com", "foo@bar.com"},
		},
		{
			in:   "foo@bar.com\r\nbaz@bar.com\r\n",
			want: []string{"baz@bar.com", "foo@bar.com"},
		},
		{
			in:   "foo@bar.com\rbaz@bar.com\r",
			want: []string{"baz@bar.com", "foo@bar.com"},
		},
		{
			in:   "# foo@bar.com\nbaz@bar.com\nzab@zab.com # comment",
			want: []string{"baz@bar.com", "zab@zab.com"},
		},
		{
			in:   "# foo@bar.com\nbaz@bar.com\n# comment\nzab@zab.com\n",
			want: []string{"baz@bar.com", "zab@zab.com"},
		},
	} {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			r := Unmarshal([]byte(tc.in))
			sort.Strings(tc.want)

			got := r.IDs()
			sort.Strings(got)

			assert.Equal(t, tc.want, got, tc.in)
		})
	}
}

func TestEndToEnd(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		op   func(r *Recipients) error
		out  string
	}{
		{
			name: "simple",
			in:   "0xDEADBEEF",
			op: func(r *Recipients) error {
				r.Add("0xFEEDBEEF")

				return nil
			},
			out: "0xDEADBEEF\n0xFEEDBEEF\n",
		},
		{
			name: "comments",
			in: `0xDEADBEEF # john doe

# some disabled ones
# 0xFOOBAR

0xFEEDBEEF
`,
			op: func(r *Recipients) error {
				r.Remove("0xFEEDBEEF")

				return nil
			},
			out: `0xDEADBEEF # john doe

# some disabled ones
# 0xFOOBAR

`,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := Unmarshal([]byte(tc.in))
			assert.NoError(t, tc.op(r))
			buf := r.Marshal()
			assert.Equal(t, tc.out, string(buf))
		})
	}
}
