package trygo

import (
	"sync/atomic"
)

type statinfo struct {
	app *App
	//当前并发连接数量
	concurrentConns int32
	//最大并发连接峰值
	peakConcurrentConns int32
}

func newStatinfo(app *App) *statinfo {
	return &statinfo{app: app}
}

func (s *statinfo) incConcurrentConns() {
	conns := atomic.AddInt32(&s.concurrentConns, 1)
	if conns > atomic.LoadInt32(&s.peakConcurrentConns) {
		atomic.StoreInt32(&s.peakConcurrentConns, conns)
	}
}

func (s *statinfo) decConcurrentConns() {
	atomic.AddInt32(&s.concurrentConns, -1)
}

func (s *statinfo) ConcurrentConns() int32 {
	return atomic.LoadInt32(&s.concurrentConns)
}

func (s *statinfo) PeakConcurrentConns() int32 {
	return atomic.LoadInt32(&s.peakConcurrentConns)
}
