package metrics

import (
	"os"
	"github.com/orcaman/concurrent-map"
	"bytes"
	"net/http"
)

var DefaultRegistry Registry = Registry{os.Args[0], cmap.New()}

type Registry struct {
	prefix string
	metrics cmap.ConcurrentMap
}

func NewRegistry(prefix string) *Registry {
	return &Registry{
		prefix,
		cmap.New(),
	}
}

func (r *Registry) Register(m... Metric) {
	for _, metric := range m {
		if !r.metrics.SetIfAbsent(metric.Name(), metric) {
			panic("Cannot register metric: " + metric.Name() + "... Already exists!")
		}
	}
}

func (r *Registry) Encode(me MetricEncoder) string {
	b := make([]byte, 0)
	buf := bytes.NewBuffer(b)
	r.metrics.IterCb(func(key string, v interface{}) {
		buf.WriteString(v.(Metric).Encode(me))
	})

	return buf.String()
}



func Register(m... Metric) {
	for _, metric := range m {
		if !DefaultRegistry.metrics.SetIfAbsent(metric.Name(), metric) {
			panic("Cannot register metric: " + metric.Name() + "... Already exists!")
		}
	}
}

//func (r *Registry) Encode(me MetricEncoder) string {
//	b := make([]byte, 0)
//	buf := bytes.NewBuffer(b)
//	r.metrics.IterCb(func(key string, v interface{}) {
//		buf.WriteString(v.(Metric).Encode(me))
//		buf.WriteString("\n")
//
//	})
//	return strings.TrimRight(buf.String(), "\n")
//}

func DefaultHttpExport(me MetricEncoder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(DefaultRegistry.Encode(me)))
	})
}

func HttpExport(r *Registry, me MetricEncoder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(r.Encode(me)))
	})
}