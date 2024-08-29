package terminal

import (
	"bytes"
	"sync"
)

type PtyTerminal struct {
	lock   sync.Mutex
	size   WindowSize
	rawBuf bytes.Buffer
	lines  [][]byte
}

func (t *PtyTerminal) Consume(p []byte) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.rawBuf.Write(p)
}

func (t *PtyTerminal) Resize(size WindowSize) {
	t.size = size
}

func (t *PtyTerminal) PredictCommand() string {
	return "ls"
}
