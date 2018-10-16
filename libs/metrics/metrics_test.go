package metrics

import (
	"fmt"
	"net/http"
	"testing"
)

func TestMetrics(t *testing.T) {
	//var pe MetricEncoder
	//pe = &PrometheusEncodeur{}
	pe := &PrometheusEncodeur{}
	sm := NewSimpleMetric("test")
	fmt.Println(sm.Encode(pe))
	sm.Set(80)
	fmt.Println(sm.Encode(pe))

	tm := NewTaggedMetrics("test2", []string{"tag", "tag2"})
	tm.Get("test", "test2").Set(80)
	tm.Get("test", "test213")
	fmt.Println(tm.Encode(pe))

	Register(sm)
	Register(tm)

	http.Handle("/test", DefaultHttpExport(pe))
	http.ListenAndServe(":8080", nil)

}
