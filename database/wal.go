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
