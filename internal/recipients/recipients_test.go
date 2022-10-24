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

			sort.Strings(tc.in)
			got := string(Marshal(tc.in))
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
	} {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			got := Unmarshal([]byte(tc.in))
			sort.Strings(got)
			assert.Equal(t, tc.want, got, tc.in)
		})
	}
}
