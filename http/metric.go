package gorgonzola

import (
	"bytes"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Metric interface {
	GetName() string
	Stringify() bytes.Buffer
}

//
type Metrics struct {
	sync.RWMutex
	items   map[string]*Metric
	status  TTLList
	expires time.Duration
	name    string
}

func NewMetrics(name string) *Metrics {
	return &Metrics{
		name:    name,
		expires: time.Second * 2,
		items:   make(map[string]*Metric),
		status:  NewTTLList(),
	}
}

func (m *Metrics) Register(metric Metric, reap time.Duration) Metric {
	m.Lock()
	defer m.Unlock()
	m.items[metric.GetName()] = &metric
	go m.reap(metric.GetName(), reap)
	return metric
}

func (m *Metrics) Get(name string) Metric {
	if v, ok := m.items[name]; ok {
		return *v
	}
	return nil
}

type MetricCreator func(name string) Metric

func (metrics *Metrics) GetOrCreate(name string, f MetricCreator) Metric {
	if m := metrics.Get(name); m != nil {
		return m
	} else {
		return metrics.Register(f(name), time.Second*1)
	}
}

func (m *Metrics) reap(name string, reap time.Duration) {
	for {
		if metric, ok := m.items[name]; ok {
			m.status.Add((*metric).Stringify(), m.expires)
			time.Sleep(reap)
		}
	}
}

// page renders the metrics page
func (m *Metrics) page(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(200)

	m.status.Lock()
	defer m.status.Unlock()

	e := m.status.list.Front()
	for {
		if e != nil {
			buf := e.Value.(*TTLItem).item.(bytes.Buffer)
			w.Write(buf.Bytes())
			e = e.Next()
		} else {
			break
		}
	}

	log.Println("size: " + strconv.Itoa(m.status.list.Len()))
}
