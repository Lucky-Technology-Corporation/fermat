package main

type RingBuffer struct {
	data   []string
	max    int
	cursor int
	full   bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]string, size),
		max:  size,
	}
}

func (rb *RingBuffer) Add(item string) {
	rb.data[rb.cursor] = item
	rb.cursor = (rb.cursor + 1) % rb.max
	if rb.cursor == 0 {
		rb.full = true
	}
}

func (rb *RingBuffer) Get() []string {
	if rb.full {
		return rb.data
	}
	return rb.data[:rb.cursor]
}
