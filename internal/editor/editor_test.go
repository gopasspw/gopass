package editor

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVimArgs(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("not supported")
	}

	for ed, args := range map[string][]string{
		"vi":     {"-c", "autocmd BufNewFile,BufRead /dev/shm/gopass* setlocal noswapfile nobackup noundofile viminfo=\"\"", "-i", "NONE", "-n"},
		"nvi":    {"-c", "autocmd BufNewFile,BufRead /dev/shm/gopass* setlocal noswapfile nobackup noundofile viminfo=\"\""},
		"neovim": {"-c", "autocmd BufNewFile,BufRead /dev/shm/gopass* setlocal noswapfile nobackup noundofile shada=\"\"", "-i", "NONE", "-n"},
		"vim":    {"-c", "autocmd BufNewFile,BufRead /dev/shm/gopass* setlocal noswapfile nobackup noundofile viminfo=\"\"", "-i", "NONE", "-n"},
	} {
		assert.Equal(t, vimOptions(ed), args, ed)
	}
}
