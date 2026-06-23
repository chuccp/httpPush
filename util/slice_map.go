package util

import (
	"io"
)

const mapSmallBufferSize = 4

type kv[V any] struct {
	key   string
	value V
}

type SliceMap[V any] struct {
	buf      []*kv[V]
	off      int
	lastRead readOp
	v        V
}

func (b *SliceMap[V]) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

func (b *SliceMap[V]) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.lastRead = opInvalid
}
func (b *SliceMap[V]) Len() int { return len(b.buf) - b.off }
func (b *SliceMap[V]) grow(n int) int {
	m := b.Len()
	if m == 0 && b.off != 0 {
		b.Reset()
	}
	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if b.buf == nil && n <= mapSmallBufferSize {
		b.buf = make([]*kv[V], n, mapSmallBufferSize)
		return 0
	}
	c := cap(b.buf)
	if n <= c/2-m {
		copy(b.buf, b.buf[b.off:])
	} else if c > maxInt-c-n {
		panic(ErrTooLarge)
	} else {
		b.buf = growSliceMap(b.buf[b.off:], b.off+n)
	}
	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

func growSliceMap[V any](b []*kv[V], n int) []*kv[V] {
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	c := len(b) + n
	if c < 2*cap(b) {
		c = 2 * cap(b)
	}
	b2 := append([]*kv[V](nil), make([]*kv[V], c)...)
	copy(b2, b)
	return b2[:len(b)]
}
func (b *SliceMap[V]) Empty() bool {
	return b.empty()
}
func (b *SliceMap[V]) empty() bool { return len(b.buf) <= b.off }
func (b *SliceMap[V]) Read() (any, error) {
	if b.empty() {
		b.Reset()
		return 0, io.EOF
	}
	c := b.buf[b.off]
	b.off++
	b.lastRead = opRead
	return c, nil
}
func (b *SliceMap[V]) Each(f func(string, V)) {
	cs := b.buf[b.off:]
	for _, c := range cs {
		f(c.key, c.value)
	}
}
func (b *SliceMap[V]) EachIndex(f func(int, string, V)) {
	cs := b.buf[b.off:]
	for index, c := range cs {
		f(index, c.key, c.value)
	}
}
func (b *SliceMap[V]) Get(key string) (V, bool) {
	if b.empty() {
		return b.v, false
	}
	cs := b.buf[b.off:]
	for _, c := range cs {
		if c.key == key {
			return c.value, true
		}
	}
	return b.v, false
}
func (b *SliceMap[V]) Delete(key string) {
	cs := b.buf[b.off:]
	for index, c := range cs {
		if c.key == key {
			if index > 0 {
				copy(cs[1:], cs[:index])
			}
			b.off++
			b.lastRead = opRead
		}
	}
	if b.empty() {
		b.Reset()
	}
}

func (b *SliceMap[V]) Put(key string, value V) error {
	return b.write(&kv[V]{key: key, value: value})
}

func (b *SliceMap[V]) PutOrReplace(key string, value V) error {
	cs := b.buf[b.off:]
	for index, c := range cs {
		if c.key == key {
			cs[index].value = value
			return nil
		}
	}
	return b.Put(key, value)
}

func (b *SliceMap[V]) write(c *kv[V]) error {
	b.lastRead = opInvalid
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		m = b.grow(1)
	}
	b.buf[m] = c
	return nil
}
