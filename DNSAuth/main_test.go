package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Packet-Clearing-House/DNSAuth/libs/metrics"
)

var res = `dnsauth_queries,direction=R,pop=xyz,qtype=NS,rcode=noerror,customer=Unknown,zone=Unknown,protocol=tcp,version=4 value=1 1508260200000000000
dnsauth_queries,direction=Q,pop=xyz,qtype=NS,rcode=none,customer=Unknown,zone=Unknown,protocol=udp,version=6 value=1 1508260080000000000
`

func TestCounters(t *testing.T) {
	limiter := make(chan bool)
	close(limiter)

	metrics.DefaultRegistry.Register(dnsqueries)
	cfg, _ := LoadConfig("./tests/dnsauth.toml")
	customerDB = NewCustomerDB("")
	// todo - this test file doesn't exist any more, need to create a dummy one with this name
	aggregate("./tests/mon-01.xyz.foonet.net_2017-10-17.17-07.dmp.gz", limiter, cfg)

	str := metrics.DefaultRegistry.Encode(&metrics.InfluxEncodeur{})

	// #WORST CHECK EVER!
	assert.Equal(t, len(str), len(res), "Wrong result length")
}
