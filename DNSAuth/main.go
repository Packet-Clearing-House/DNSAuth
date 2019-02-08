package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Packet-Clearing-House/DNSAuth/libs/dnsdist"
	"github.com/Packet-Clearing-House/DNSAuth/libs/metrics"
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

var confpath = flag.String("c", "./dnsauth.toml", "Path for the config path (default is ./dnsauth.toml")

var dnsqueries = metrics.NewTTLTaggedMetrics("dnsauth_queries", []string{"direction", "pop", "qtype", "rcode", "customer", "zone", "protocol", "version", "prefix", "origin_as"}, 500)
var customerDB *CustomerDB

func main() {

	flag.Parse()

	log.Printf("Loading config file %s...\n", *confpath)
	config, err := LoadConfig(*confpath)
	if err != nil {
		log.Fatalln("FAILED: ", err)
	}
	log.Println("OK!")

	DB_URL = config.CustomerDB
	INFLUX_URL = config.InfluxDB

	// Starting the customerDB fetching process
	log.Println("Initializing customer DB (will be refresh every " + strconv.Itoa(config.CustomerRefresh) + " hours)...")
	customerDB = NewCustomerDB(DB_URL)
	go func() {
		// Refresh function
		refresh := func() {
			log.Println("[CustomerDB] Refreshing list from mysql...")
			if err := customerDB.Refresh(); err != nil {
				log.Println("[CustomerDB] ERROR: Could not refresh customer list (", err, ")!")
			}
		}

		refresh()
		for _ = range time.Tick(time.Duration(config.CustomerRefresh) * time.Hour) {
			refresh()
		}
	}()

	// Running the metric pushing process
	metrics.DefaultRegistry.Register(dnsqueries)
	go func() {
		for {
			push(&metrics.DefaultRegistry)
			time.Sleep(time.Minute)
		}
	}()

	limiter := make(chan bool, 20)
	files := make(map[string]interface{})
	newFiles := make(map[string]interface{})

	visit := func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".dmp.gz") {

			if _, found := files[path]; !found {
				go aggregate(path, limiter, config)
				limiter <- true
			}
			newFiles[path] = true
		}
		return nil
	}

	err = filepath.Walk(config.WatchDir, func(path string, f os.FileInfo, err error) error {
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

func aggregate(filePath string, limiter chan bool, config *Config) {

	starttime := time.Now()

	defer func() { <-limiter }()

	fileHandle, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer fileHandle.Close()

	reader, err := gzip.NewReader(fileHandle)
	if err != nil {
		log.Println(filePath, ": ", err)
		return
	}
	defer reader.Close()

	index := strings.LastIndex(filePath, "mon-") + len("mon-")
	mon := filePath[index : index+2]
	pop := filePath[index+3 : index+6]

	index = strings.LastIndex(filePath, "_") + 1
	timestamp := filePath[index : index+16]

	date, err := time.Parse(LAYOUT, timestamp)
	if err != nil {
		log.Println(filePath, ": ", err)
		return
	}

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
			log.Println("Issue unformatting line:", line, " for dump ", filePath)
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
	err = cleanupFile(filePath, config)
	if err != nil {
		log.Printf("Failed to clean up %s. Reason: %s", filePath, err)
	}
	proctime := time.Since(starttime)
	log.Printf("Processed dump [mon-%s-%s](%s - %s): %d lines in (%s) seconds!\n",
		mon, pop, initialdate, date, cpt, proctime)

}

func handleQuery(time time.Time, pop, line string) {

	fields := strings.Fields(line)

	prefix := ""
	originAs := ""
	version := "4"

	// Resolving destination address to client
	qname := fields[QNAME][:len(fields[QNAME])-1]
	zone, name := customerDB.Resolve(qname)

	if ipv := net.ParseIP(fields[CLIENT_IP]); ipv != nil {
		if ipv.To4() == nil {
			version = "6"
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

	dnsqueries.GetAt(time, fields[DIRECTION], pop, qtypestr, rcodestr, name, zone, protocol, version, originAs, prefix).Inc()
}

func cleanupFile(filePath string, config *Config) error {
	var err error
	switch config.CleanupAction {
	case "move":
		err = cleanupFileMove(filePath, config.CleanupDir)
	case "delete":
		err = cleanupFileDelete(filePath)
	case "none":
	default:
		err = fmt.Errorf("Invalid config setting for cleanup action: %s", config.CleanupAction)
	}
	return err
}

func cleanupFileMove(filePath string, destDir string) error {
	log.Printf("Moving file %s to %s\n", filePath, destDir)
	return os.Rename(filePath, destDir+"/"+filepath.Base(filePath))
}

func cleanupFileDelete(filePath string) error {
	log.Printf("Removing file %s\n", filePath)
	return os.Remove(filePath)
}
