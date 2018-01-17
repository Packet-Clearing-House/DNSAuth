package main

import (
	"log"
	"stash.pch.net/rrdns/DNSAuth/libs/metrics"
	"strings"
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
)

var INFLUX_URL = "http://127.0.0.1:8086/write?db=authdns"

func push(registry *metrics.Registry) {
	str := registry.Encode(&metrics.InfluxEncodeur{})
	splits := strings.Split(str, "\n")

	buffer := bytes.NewBuffer(nil)

	var cpt = 0
	for i, value := range splits {
		cpt += 1
		buffer.WriteString(value + "\n")
		if i % 5000 == 0  && i != 0 {
			resp, err := http.Post(INFLUX_URL, "application/octet-stream", buffer)
			if err != nil {
				log.Println(err)
			} else if resp.StatusCode != 204 {
				buf, _ := ioutil.ReadAll(resp.Body)
				log.Println(string(buf))
				resp.Body.Close()
			}
		}
	}
	resp, err := http.Post(INFLUX_URL, "application/octet-stream", buffer)
	if err != nil {
		log.Println(err)
	} else if resp.StatusCode != 204 {
		buf, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(buf))
		resp.Body.Close()
	}
	log.Println("Influx pusher inserted " + strconv.Itoa(cpt) + " points!")
}
