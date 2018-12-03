package main

import (
	"log"
	"testing"

	"github.com/Packet-Clearing-House/DNSAuth/libs/metrics"
)

var res = `dnsaut_queries{direction="Q",pop=".wo",qtype="DNSKEY",rcode="none"} 1 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="A",rcode="noerror"} 154 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="AAAA",rcode="nxdomain"} 1 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="SPF",rcode="noerror"} 1 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="SRV",rcode="noerror"} 2 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="SOA",rcode="noerror"} 2 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="PTR",rcode="noerror"} 118 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="TXT",rcode="none"} 7 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="OTHER",rcode="noerror"} 2 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="MX",rcode="noerror"} 8 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="SRV",rcode="none"} 2 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="AAAA",rcode="none"} 54 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="NS",rcode="nxdomain"} 3 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="PTR",rcode="nxdomain"} 35 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="NS",rcode="none"} 6 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="SPF",rcode="none"} 1 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="MX",rcode="none"} 8 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="TXT",rcode="noerror"} 7 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="A",rcode="none"} 163 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="A",rcode="nxdomain"} 10 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="AAAA",rcode="noerror"} 53 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="PTR",rcode="none"} 153 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="DS",rcode="none"} 18 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="NS",rcode="noerror"} 3 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="OTHER",rcode="none"} 2 1508260020000
dnsaut_queries{direction="R",pop=".wo",qtype="DS",rcode="noerror"} 17 1508260020000
dnsaut_queries{direction="Q",pop=".wo",qtype="SOA",rcode="none"} 2 1508260020000
`

func TestCounters(t *testing.T) {
	limiter := make(chan bool)
	close(limiter)

	//tree := nradix.NewTree(0)

	metrics.DefaultRegistry.Register(dnsqueries)
	cfg, _ := LoadConfig("")
	// todo - this test file doesn't exist any more, need to create a dummy one with this name
	aggregate("./tests/mon-01.xyz.foonet.net_2017-10-17.17-07.dmp.gz", limiter, cfg)

	str := metrics.DefaultRegistry.Encode(&metrics.InfluxEncodeur{})

	log.Println(str)
	// push(&metrics.DefaultRegistry)

	// #WORST CHECK EVER!
	if len(str) != len(res) {
		t.Fatal("Not getting the right result!")
	}
}
