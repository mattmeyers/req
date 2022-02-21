package cli

type history struct {
	values []string
	idx    int
	cap    int
}

func newHistory(size int) *history {
	if size < 0 {
		panic("history cannot have negative size")
	}

	return &history{
		values: make([]string, size),
		cap:    size,
		idx:    size - 1,
	}
}

func (h *history) append(s string) {
	h.idx = (h.idx + 1) % h.cap
	h.values[h.idx] = s
}

func (h *history) get(offset uint) string {
	return h.values[((h.idx-int(offset))%h.cap+h.cap)%h.cap]
}
