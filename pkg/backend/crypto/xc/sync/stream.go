package sync

import (
	"bytes"
	"fmt"
	"io"
)

type chunk struct {
	num int
	buf []byte
}

type Stream struct {
	offset int
	in     []chan chunk
	out    []chan chunk
	errc   chan error
}

func New(slots, size int) *Stream {
	s := &Stream{
		offset: 0,
		in:     make([]chan chunk, slots),
		out:    make([]chan chunk, slots),
		errc:   make(chan error, slots),
	}
	for i := range s.in {
		s.in[i] = make(chan chunk, size)
		s.out[i] = make(chan chunk, size)
	}
	return s
}

func (s *Stream) Write(num int, buf []byte) {
	if len(s.in) < 1 {
		return
	}
	slot := num % len(s.in)
	//fmt.Printf("Write: %d -> %d\n", num, slot)
	pl := chunk{
		num: num,
		buf: make([]byte, len(buf)),
	}
	copy(pl.buf, buf)
	s.in[slot] <- pl
}

func (s *Stream) Close() {
	for _, in := range s.in {
		close(in)
	}
}

func (s *Stream) Work(fn func(int, []byte, io.Writer) error) error {
	if len(s.in) != len(s.out) {
		return fmt.Errorf("misconfiguration")
	}
	for i := 0; i < len(s.in); i++ {
		go s.doWork(fn, s.in[i], s.out[i], i)
	}
	return nil
}

func (s *Stream) doWork(fn func(int, []byte, io.Writer) error, in chan chunk, out chan chunk, slot int) {
	ob := &bytes.Buffer{}
	for pl := range in {
		ob.Reset()
		if err := fn(pl.num, pl.buf, ob); err != nil {
			s.errc <- err
			continue
		}
		plOut := chunk{
			num: pl.num,
			buf: make([]byte, ob.Len()),
		}
		copy(plOut.buf, ob.Bytes())
		out <- plOut
	}
	close(out)
}

func (s *Stream) Consume(fn func(int, []byte) error) error {
	if len(s.out) < 1 {
		return nil
	}
	offset := 0
	slot := 0
	for {
		select {
		case err := <-s.errc:
			return err
		default:
		}
		if slot >= len(s.out) {
			//fmt.Printf("Slot: %d - len(out): %d\n", slot, len(s.out))
			panic("invalid slot number")
		}
		pl, ok := <-s.out[slot]
		if !ok {
			return nil
		}
		if offset != pl.num {
			return fmt.Errorf("reordering error (expected: %d - got: %d)", s.offset, pl.num)
		}
		if err := fn(pl.num, pl.buf); err != nil {
			return err
		}
		slot = (slot + 1) % len(s.out)
		offset++
	}
}
