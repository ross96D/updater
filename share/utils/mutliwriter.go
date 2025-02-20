// Multiwriter that does not error on a single error
package utils

import "io"

// multiwriter that ignore all errors
type multiwriter struct {
	writers SwapSlice[io.Writer]
}

func (m *multiwriter) Write(p []byte) (n int, err error) {
	if m.writers.Len() == 0 {
		return 0, io.EOF
	}

	delete := make([]int, 0)
	for i, w := range m.writers.Slice() {
		n, err = w.Write(p)
		if err != nil {
			delete = append(delete, i)
			continue
		}
		if n != len(p) {
			delete = append(delete, i)
			continue
		}
	}
	for _, i := range delete {
		m.writers.Remove(i)
	}
	return len(p), nil
}

func MultiWriter(writers ...io.Writer) io.Writer {
	allwriters := SwapSlice[io.Writer]{}
	allwriters.Append(writers...)
	return &multiwriter{writers: allwriters}
}
