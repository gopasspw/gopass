package sync

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStream(t *testing.T) {
	for _, slots := range []int{0, 1, 4, 32} {
		fmt.Printf("Slots: %d\n", slots)
		for _, size := range []int{0, 1, 1024} {
			fmt.Printf(" Size: %d\n", size)
			for _, blocks := range []int{0, 1, 512} {
				fmt.Printf("  Blocks: %d\n", blocks)
				s := New(slots, size)

				go func() {
					for i := 0; i < blocks; i++ {
						buf := make([]byte, 1024)
						rand.Read(buf)
						s.Write(i, buf)
					}
					s.Close()
				}()

				require.NoError(t, s.Work(func(num int, buf []byte) ([]byte, error) {
					for i := 0; i < int(rand.Int31n(1024*1024*10)); i++ {
						_ = sha512.New().Sum(buf)
					}
					return buf, nil
				}))

				offset := 0
				require.NoError(t, s.Consume(func(num int, buf []byte) error {
					assert.Equal(t, offset, num)
					//fmt.Printf("Offset: %d - Num: %d\n", offset, num)
					offset++
					return nil
				}))
			}
		}
	}
}
