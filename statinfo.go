package trygo

import (
	"reflect"
	"sync"
	"sync/atomic"
)

var atomicInt64Enable bool

func init() {
	atomicInt64Enable = reflect.TypeOf(int(0)).Bits() >= 64
}

type statinfo struct {
	app *App
	//当前并发连接数量
	concurrentConns int32
	//最大并发连接峰值
	peakConcurrentConns int32
	//当前进行中的请求数量
	currentRequests int64
	//累计请求数量
	totalRequests int64
	reqsLocker    sync.RWMutex
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

func (s *statinfo) incCurrentRequests() {
	if atomicInt64Enable {
		atomic.AddInt64(&s.currentRequests, 1)
		atomic.AddInt64(&s.totalRequests, 1)
	} else {
		s.reqsLocker.Lock()
		defer s.reqsLocker.Unlock()
		s.currentRequests++
		s.totalRequests++
	}
}

func (s *statinfo) decCurrentRequests() {
	if atomicInt64Enable {
		atomic.AddInt64(&s.currentRequests, -1)
	} else {
		s.reqsLocker.Lock()
		defer s.reqsLocker.Unlock()
		s.currentRequests--
	}
}

func (s *statinfo) CurrentRequests() int64 {
	if atomicInt64Enable {
		return atomic.LoadInt64(&s.currentRequests)
	} else {
		s.reqsLocker.RLock()
		defer s.reqsLocker.RUnlock()
		return s.currentRequests
	}
}

func (s *statinfo) TotalRequests() int64 {
	if atomicInt64Enable {
		return atomic.LoadInt64(&s.totalRequests)
	} else {
		s.reqsLocker.RLock()
		defer s.reqsLocker.RUnlock()
		return s.totalRequests
	}
}
