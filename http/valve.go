package gorgonzola

import (
	"bytes"
	"strconv"
	"sync"
	"time"
)

type ValveValue struct {
	sync.RWMutex
	Metric
	vv *ValveVec
	v  float64
}

type ValveVec struct {
	sync.RWMutex
	name     string
	dirty    bool
	lastRead time.Time
	labels   map[string]string
	value    map[string]*ValveValue
}

func NewValveVec(name string) *ValveVec {
	return &ValveVec{
		name:     name,
		value:    make(map[string]*ValveValue),
		lastRead: time.Now(),
	}
}

func (vv *ValveVec) GetName() string {
	return vv.name
}

func (vv *ValveVec) Add(name string, value float64) {
	if valve, ok := vv.value[name]; ok {
		valve.Add(value)
	} else {
		vv.Lock()
		defer vv.Unlock()
		vv.dirty = true
		vv.value[name] = &ValveValue{vv: vv, v: value}
	}
}

func (vv *ValveVec) Stringify() bytes.Buffer {
	var buffer bytes.Buffer

	vv.Lock()
	now := time.Now()
	nowUnix := now.UnixNano() // keep it stable for this method

	for k, x := range vv.value {
		buffer.WriteString(vv.name + ",from=" + k + " value=" + strconv.FormatFloat(x.v, 'f', -1, 64) + " " + strconv.FormatInt(nowUnix, 10) + "\n")
		delete(vv.value, k)
	}
	vv.lastRead = now
	vv.Unlock()
	return buffer
}

func (value *ValveValue) Add(delta float64) {
	value.Lock()
	defer value.Unlock()
	value.vv.dirty = true
	value.v += delta
}
