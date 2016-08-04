package gorgonzola

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type GoMetrics struct {
	sync.RWMutex
	Name string
	mem  float64
}

func (vv *GoMetrics) GetName() string {
	return vv.Name
}

func (x *GoMetrics) Stringify() bytes.Buffer {
	var buffer bytes.Buffer
	nowUnix := time.Now().UnixNano()
	ts := strconv.FormatInt(nowUnix, 10)
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	buffer.WriteString("memory_alloc value=" + strconv.FormatUint(m.Alloc, 10) + " " + ts + "\n")
	buffer.WriteString("number_of_goroutines value=" + strconv.Itoa(runtime.NumGoroutine()) + " " + ts + "\n")

	return buffer
}
