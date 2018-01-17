package main

import (
	"log"
	"time"
	"flag"
	"os"
	"strings"
	"path/filepath"
	"compress/gzip"
	"bytes"
	"bufio"
	"stash.pch.net/rrdns/DNSAuth/libs/metrics"
	"strconv"
	"stash.pch.net/rrdns/DNSAuth/libs/dnsdist"
	"github.com/asergeyev/nradix"
	"stash.pch.net/rrdns/DNSAuth/DNSAuth/bgp"
)

const (
	DIRECTION = iota
	CLIENT_IP
	NS_IP
	PROTOCOL
	OPCODE
	QTYPE
	QNAME
	PACKET_SIZE
	RCODE
)

const LAYOUT = "2006-01-02.15-04"

var confpath = flag.String("c", ".", "Path for the config path (default is ./dnsauth.toml")


var dnsqueries = metrics.NewTTLTaggedMetrics("dnsauth_queries", []string{"direction", "pop", "qtype", "rcode", "customer", "protocol", "prefix", "origin_as"}, 500)
var tree *nradix.Tree

func main() {
	
	flag.Parse()

	log.Println("Loading config file...")
	config, err := LoadConfig(*confpath)
	if err != nil {
		log.Println("Error loading config file: ", err)
	}
	log.Println("OK!")
	
	DB_URL = config.CustomerDB
	INFLUX_URL = config.InfluxDB

	log.Println("Getting customer list from postgres...")
	t, err := getCustomerTree()
	tree = t
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("OK!")


	if config.BGP != nil {
		log.Println("Starting BGP Resolver...")
		err = bgp.Start(*config.BGP)
		if err != nil {
			log.Fatalln("Failed to start BGP Server: ", err)
		}
	}
	log.Println("OK!")
	//http.Handle(
	//	"/metrics",
	//	metrics.HttpExport(&metrics.DefaultRegistry, &metrics.PrometheusEncodeur{}))
	//
	//go http.ListenAndServe(":8080", nil)


	metrics.DefaultRegistry.Register(dnsqueries)
	go func() {
		for {
			log.Println("Pushing metrics!!")
			starttime := time.Now()
			push(&metrics.DefaultRegistry)
			proctime := time.Since(starttime)
			log.Println("Took " + proctime.String() + "seconds")
			time.Sleep(time.Minute)
		}
	}()



	limiter := make(chan bool, 20)
	files := make(map[string]interface{})
	newFiles := make(map[string]interface{})

	visit := func (path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".dmp.gz") {

			if _, found := files[path]; found {
				newFiles[path] = true
			} else {
				newFiles[path] = true
				go aggreagate(path, limiter)
				limiter <-true
			}
		}
		return nil
	}

	err = filepath.Walk(config.WatchDir, func (path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".dmp.gz") {
			files[path] = true
		}
		return nil
	})
	if err != nil {
		log.Fatalln("Error while watching directory: ", err)
	}

	for {
		time.Sleep(time.Second * 30)
		err = filepath.Walk(config.WatchDir, visit)
		if err != nil {
			log.Fatal(err)
		}
		files = newFiles
		newFiles = make(map[string]interface{})
	}
}

func aggreagate(filepath string, limiter chan bool) {

	starttime := time.Now()

	defer func() {<-limiter}()

	fileHandle, err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		return
	}
	defer fileHandle.Close()

	reader, err := gzip.NewReader(fileHandle)
	if err != nil {
		log.Println(filepath, ": ", err)
		return
	}
	defer reader.Close()

	index := strings.LastIndex(filepath,"mon-") + len("mon-")
	mon := filepath[index:index+2]
	pop := filepath[index+3:index+6]

	index = strings.LastIndex(filepath,"net_") + len("net_")
	timestamp := filepath[index:index+16]

	date, _ := time.Parse(LAYOUT, timestamp)

	buffer := bytes.NewBuffer(nil)
	cpt := 0
	fileScanner := bufio.NewScanner(reader)
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if line[0] == 'Q' {
			line += " -1"
		} else if line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 9 {
			log.Println("Issue unformatting line:", line, " for dump ", filepath)
			continue
		}
		buffer.WriteString(line)
		buffer.WriteString("\n")
		cpt += 1
	}

	if cpt == 0 {
		return
	}

	interval := 60 * 3 * 1000000.0 / uint(cpt)
	date, e := time.Parse(LAYOUT, timestamp)
	if e != nil {
		log.Fatalln(e)
	}
	initialdate := date


	for {
		date = date.Add(time.Duration(interval) * time.Microsecond)

		line, err := buffer.ReadString('\n')
		if err != nil {
			break
		}

		handleQuery(date.Truncate(time.Minute), pop, line)

	}
	proctime := time.Since(starttime)
	log.Printf("Processed dump [mon-%s-%s](%s - %s): %d lines in (%s) seconds!\n",
		mon, pop, initialdate, date, cpt, proctime)

}


func handleQuery(time time.Time, pop, line string) {

	fields := strings.Fields(line)

	name := "Unknown"
	prefix := ""
	originAs := ""

	// Resolving destination address to client
	c, _ := tree.FindCIDR(fields[NS_IP])

	// If we do find a result...
	if c != nil {
		customer := c.(*Customer)
		name = customer.Name

		// ...resolving client ip through BGP
		if customer.PrefixMonit || customer.ASNMonit {
			entry, err := bgp.Resolve(fields[CLIENT_IP])
			if err == nil {
				// I SHOULD DO SOMETHING HERE #DEBUG?
				originAs = strconv.Itoa(int(entry.Path[len(entry.Path) - 1]))
				if customer.PrefixMonit {
					prefix = entry.Prefix
				}
			}
		}
	}

	rcode, err := strconv.Atoi(fields[RCODE])
	if err != nil {
		log.Fatalln(err)
	}
	qtype, err := strconv.Atoi(fields[QTYPE])
	if err != nil {
		log.Fatalln(err)
	}

	protocol := "udp"
	if fields[PROTOCOL] == "0" {
		protocol = "tcp"
	}

	rcodestr := dnsdist.RCode(rcode).String()
	qtypestr := dnsdist.QType(qtype).String()
	if rcode == -1 {
		rcodestr = "none"
	}

	dnsqueries.GetAt(time, fields[DIRECTION], pop, qtypestr, rcodestr, name, protocol, originAs, prefix).Inc()
}