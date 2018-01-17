package metrics

import (
	"strings"
	"strconv"
	"bytes"
)

type PrometheusEncodeur struct {}


func (pe *PrometheusEncodeur) EncodeSimpleMetric(sm *SimpleMetric) string {
	date := ""
	if sm.time != nil {
		date = strconv.FormatInt(sm.time.Unix(), 10)
	}
	return sm.Name() + " " + strconv.FormatUint(sm.Value(), 10) + date + "\n"
}

func (pe *PrometheusEncodeur) EncodeTaggedMetrics(tm *TaggedMetrics) string {
	buf := bytes.NewBuffer(nil)

	tm.Values.IterCb(func(key string, v interface{}) {
		taggedValues := strings.Split(key, "\n")
		//
		buf.WriteString(tm.Name())
		buf.WriteString("{")
		if taggedValues[0] != "" {
			buf.WriteString(tm.Tags[0])
			buf.WriteString("=\"")
			buf.WriteString(taggedValues[0])
			buf.WriteString("\"")
		}
		for i, tag := range tm.Tags[1:] {
			if taggedValues[i+1] != "" {
				buf.WriteString(",")
				buf.WriteString(tag)
				buf.WriteString("=\"")
				buf.WriteString(taggedValues[i+1])
				buf.WriteString("\"")
			}
		}
		date := ""
		if v.(*SimpleMetric).time != nil {
			date = strconv.FormatInt(v.(*SimpleMetric).time.Unix(), 10)
		}
		buf.WriteString("} ")
		buf.WriteString(strconv.FormatUint(v.(*SimpleMetric).Value(), 10))
		buf.WriteString(" ")
		buf.WriteString(date)
		buf.WriteString("000000000\n")
	})
	return buf.String()
}

func (pe *PrometheusEncodeur) EncodeFuncMetric(tm *FuncMetric) string {
	return tm.Name() + " " + strconv.FormatUint(tm.Value(), 10) + "\n"
}




type InfluxEncodeur struct {}


func (pe *InfluxEncodeur ) EncodeSimpleMetric(sm *SimpleMetric) string {
	date := ""
	if sm.time != nil {
		date = strconv.FormatInt(sm.time.Unix(), 10)
	}
	return sm.Name() + ", value=" + strconv.FormatUint(sm.Value(), 10) + " " + date + "\n"
}

func (pe *InfluxEncodeur ) EncodeTaggedMetrics(tm *TaggedMetrics) string {

	buf := bytes.NewBuffer(nil)

	tm.Values.IterCb(func(key string, v interface{}) {
		taggedValues := strings.Split(key, "\n")

		buf.WriteString(tm.Name())
		buf.WriteString(",")

		if taggedValues[0] != "" {
			value := strings.Replace(taggedValues[0], " ", "\\ ", -1)
			buf.WriteString(tm.Tags[0])
			buf.WriteString("=")
			buf.WriteString(value)
		}
		for i, tag := range tm.Tags[1:] {
			if taggedValues[i+1] != "" {
				value := strings.Replace(taggedValues[i+1], " ", "\\ ", -1)
				buf.WriteString(",")
				buf.WriteString(tag)
				buf.WriteString("=")
				buf.WriteString(value)
			}
		}
		buf.WriteString(" value=")
		buf.WriteString(strconv.FormatUint(v.(*SimpleMetric).Value(), 10))
		buf.WriteString(" ")

		date := ""
		if v.(*SimpleMetric).time != nil {
			date = strconv.FormatInt(v.(*SimpleMetric).time.Unix(), 10)
		}
		buf.WriteString(date)
		buf.WriteString("000000000\n")
	})
	return buf.String()
}

func (pe *InfluxEncodeur) EncodeFuncMetric(tm *FuncMetric) string {
	return tm.Name() + ",value=" + strconv.FormatUint(tm.Value(), 10) + "000000000\n"
}