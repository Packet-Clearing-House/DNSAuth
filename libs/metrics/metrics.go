package metrics

import (
	"github.com/orcaman/concurrent-map"
	"sync/atomic"
	"time"
)



type Labels map[*string]string

func (ls Labels) Hash() string {
	hash := ""
	for _, value := range ls {
		hash += value
	}
	return hash
}

type Metrics interface {
	Name() string
	Value() uint64
}


type Vector interface {
	Name()
	Collect()
}



type Metric interface {
	Name() string
	Encode(encoder MetricEncoder) string
}

func (sm *SimpleMetric) Encode(encoder MetricEncoder) string {
	return encoder.EncodeSimpleMetric(sm)
}

type Collector interface {
	Collect()
}

type SimpleMetric struct {
	name *string
	time *time.Time
	value uint64
}

type FuncMetric struct {
	name string
	fn func() uint64
}

func NewSimpleMetric(name string) *SimpleMetric {
	return &SimpleMetric{
		&name,
		nil,
		0,
	}
}

func NewFuncMetric(name string, fn func() uint64) *FuncMetric {
	return &FuncMetric{
		name,
		fn,
	}
}

func (fm *FuncMetric) Name() string{
	return fm.name
}

func (fm *FuncMetric) Value() uint64 {
	return fm.fn()
}

func (fm *FuncMetric) Encode(encoder MetricEncoder) string {
	return encoder.EncodeFuncMetric(fm)
}

func (sm *SimpleMetric) Set(value uint64) {
	atomic.StoreUint64(&sm.value, value)
}

func (sm *SimpleMetric) Inc() {
	atomic.AddUint64(&sm.value, 1)
}

func (sm *SimpleMetric) Dec() {
	atomic.StoreUint64(&sm.value, atomic.LoadUint64(&sm.value) - 1)
}

func (sm *SimpleMetric) Value() uint64 {
	return atomic.LoadUint64(&sm.value)
}

func (sm *SimpleMetric) Name() string {
	return *sm.name
}

type MetricEncoder interface {
	EncodeSimpleMetric(sm *SimpleMetric) string
	EncodeTaggedMetrics(tm *TaggedMetrics) string
	EncodeFuncMetric(tm *FuncMetric) string
}


type TaggedMetrics struct {
	name string
	Tags []string
	Updates cmap.ConcurrentMap
	Values cmap.ConcurrentMap
}

func (tm *TaggedMetrics) GetAt(date time.Time, tags ...string) *SimpleMetric {
	if len(tags) != len(tm.Tags) {
		panic("Mismatch with tags!")
	}
	key := ""
	for _, elem := range tags {
		key += elem + "\n"
	}
	key += date.String()
	if value, ok := tm.Values.Get(key); ok {
		tm.Updates.Set(key, time.Now().Unix())
		return value.(*SimpleMetric)
	}
	sm := NewSimpleMetric(tm.name)
	sm.time = &date
	tm.Values.Set(key, sm)
	tm.Updates.Set(key, time.Now().Unix())
	return sm
}

func (tm *TaggedMetrics) Get(tags ...string) *SimpleMetric {
	if len(tags) != len(tm.Tags) {
		panic("Mismatch with tags!")
	}
	key := ""
	for _, elem := range tags {
		key += elem + "\n"
	}
	if value, ok := tm.Values.Get(key); ok {
		tm.Updates.Set(key, time.Now().Unix())
		return value.(*SimpleMetric)
	}
	sm := NewSimpleMetric(tm.name)
	tm.Values.Set(key, sm)
	tm.Updates.Set(key, time.Now().Unix())
	return sm
}


func (tm *TaggedMetrics) Name() string {
	return tm.name
}


func (tm *TaggedMetrics) Encode(encoder MetricEncoder) string {
	return encoder.EncodeTaggedMetrics(tm)
}

//func (tm *TaggedMetrics) SetConstTags(consts map[string]string) {
//	for tuple := range tm.constTags.IterBuffered() {
//		tm.constTags.Remove(tuple.Key)
//	}
//	for key, value := range consts {
//		tm.constTags.Set(key, value)
//	}
//}

func (tm *TaggedMetrics) handleQuiescentMetrics(ttl int64) {
	for {
		for tuple := range tm.Updates.IterBuffered() {
			value := tuple.Val.(int64)
			if value + ttl <= time.Now().Unix() {
				tm.Updates.Remove(tuple.Key)
				tm.Values.Remove(tuple.Key)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func NewTaggedMetrics(name string, tags []string) *TaggedMetrics {

	tm := &TaggedMetrics {
		name,
		tags,
		cmap.New(),
		cmap.New(),
	}
	return tm
}

func NewTTLTaggedMetrics(name string, tags []string, ttl int64) *TaggedMetrics {

	tm := &TaggedMetrics {
		name,
		tags,
		cmap.New(),
		cmap.New(),
	}
	if ttl > 0 {
		go tm.handleQuiescentMetrics(ttl)
	}

	return tm
}
