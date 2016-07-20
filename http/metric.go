package gorgonzola

import (
	"bytes"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type XMetric interface {
	GetName() string
	Stringify() bytes.Buffer
}

//
type Metrics struct {
	sync.RWMutex
	items  map[string]*XMetric
	status TTLDaemon
}

func NewMetrics() *Metrics {
	return &Metrics{
		items:  make(map[string]*XMetric),
		status: NewTTLDaemon(),
	}
}

func (m *Metrics) Register(metric XMetric, reap time.Duration) XMetric {
	m.Lock()
	defer m.Unlock()
	m.items[metric.GetName()] = &metric
	go m.reap(metric.GetName(), reap)
	return metric
}

func (m *Metrics) Get(name string) XMetric {
	if v, ok := m.items[name]; ok {
		return *v
	}
	return nil
}

type XMetricCreate func(name string) XMetric

func (metrics *Metrics) GetOrCreate(name string, f XMetricCreate) XMetric {
	if m := metrics.Get(name); m != nil {
		return m
	} else {
		return metrics.Register(f(name), time.Second*1)
	}
}

func (m *Metrics) reap(name string, reap time.Duration) {

	for {
		if metric, ok := m.items[name]; ok {

			ttl := &TTL{
				item: (*metric).Stringify(),
			}
			ttl.touch()

			m.status.items[RandStringBytesMaskImprSrc(16)] = ttl

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
	for _, item := range m.status.items {
		x := item.item.(bytes.Buffer)
		w.Write(x.Bytes())
	}
}

// TODO cleanup
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// ********************************************************************************

type MMetric struct {
	sync.RWMutex
	Name string
	mem  float64
}

func (vv *MMetric) GetName() string {
	return vv.Name
}

func (x *MMetric) Stringify() bytes.Buffer {
	var buffer bytes.Buffer

	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	nowUnix := time.Now().UnixNano()

	buffer.WriteString(x.Name + " value=" + strconv.FormatUint(m.Alloc, 10) + " " + strconv.FormatInt(nowUnix, 10) + "\n")

	return buffer
}
