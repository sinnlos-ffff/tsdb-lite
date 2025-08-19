package database

import (
	"bufio"
	"os"
	"time"
)

type walReq struct {
	p    Point
	done chan error
}

type WAL struct {
	ch        chan walReq
	w         *bufio.Writer
	f         *os.File
	flushTick *time.Ticker // e.g., 20ms
}

func (w *WAL) loop() {
	buf := make([]byte, 0, 1<<20)
	for {
		select {
		case req := <-w.ch:
			// TODO: Implement encoding logic
			buf = appendRecord(buf, req.p)
			if len(buf) > 64<<10 { // 64KiB batch
				w.w.Write(buf)
				buf = buf[:0]
			}

			// TODO: Consider moving ACK after w.f.Sync()
			req.done <- nil
		case <-w.flushTick.C:
			if len(buf) > 0 {
				w.w.Write(buf)
				buf = buf[:0]
			}
			w.w.Flush()
			w.f.Sync() // group commit
		}
	}
}
