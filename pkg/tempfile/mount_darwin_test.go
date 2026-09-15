//go:build darwin

package tempfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDev(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		out  string
		want string
		err  bool
	}{
		{
			// hdid pads the device node with spaces and tabs.
			name: "hdid",
			out:  "/dev/disk4          \t                               \t\n",
			want: "/dev/disk4",
		},
		{
			// diskutil emits the bare device node for a --noMount ramdisk.
			name: "diskutil",
			out:  "/dev/disk4\n",
			want: "/dev/disk4",
		},
		{
			// diskutil appends a content hint and a mount point when it has them.
			name: "diskutil with content hint",
			out:  "/dev/disk4\tGUID_partition_scheme\t/Volumes/foo\n",
			want: "/dev/disk4",
		},
		{
			name: "empty",
			out:  "",
			err:  true,
		},
		{
			name: "whitespace only",
			out:  " \t\n",
			err:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dev, err := parseDev(tc.out)
			if tc.err {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, dev)
		})
	}
}
