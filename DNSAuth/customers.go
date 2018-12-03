package main

import (
	"database/sql"
	"math/big"
	"net"
	"sync"

	"github.com/Packet-Clearing-House/DNSAuth/libs/iprange"
	atree "github.com/Packet-Clearing-House/go-datastructures/augmentedtree"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DB_URL = "root:pass@(127.0.0.1)/customers"

type customer struct {
	Name string
	Zone string
}

// Function that reverse a word (test.com -> moc.tset)
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ipBytes returns the byte array of a net.IP, supporting both IPv4 and IPv6
func ipBytes(ip net.IP) []byte {
	if ipBytes4 := ip.To4(); ipBytes4 != nil {
		return ipBytes4
	}
	return ip
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func longestCommonPrefix(strA, strB string) int {
	if strA == "" || strB == "" {
		return 0
	}
	runeA := []rune(strA)
	runeB := []rune(strB)
	maxLength := max(len(runeA), len(runeB))
	var i int
	for i < maxLength && runeA[i] == runeB[i] {
		i++
	}
	return i
}

func matchQueryPrefix(qnameRev string, intervals atree.Intervals) []*customer {
	var customers []*customer
	var longestMatch int
	for i := 0; i < len(intervals); i++ {
		interval := intervals[i]
		ipRange := interval.(*iprange.IPInterval)
		cust := ipRange.Value.(customer)
		match := longestCommonPrefix(qnameRev, cust.Zone)
		if match > longestMatch {
			customers = []*customer{&cust}
			longestMatch = match
		} else if match == longestMatch {
			customers = append(customers, &cust)
		}
	}
	return customers
}

type CustomerDB struct {
	*sync.Mutex
	dburl string
	atree *atree.Tree
}

// Resolve the customer name from DNS qname
// Returns Unknown if not found
func (c *CustomerDB) Resolve(qname string, ip net.IP) (string, string, int) {

	name := "Unknown"
	zone := "Unknown"
	var found int
	// Match by IP range
	c.Lock()
	var ipInt big.Int
	ipInt.SetBytes(ipBytes(ip))
	queryInterval := iprange.NewSingleDimensionInterval(ipInt, ipInt, 0, nil)
	intervals := (*c.atree).Query(queryInterval)
	c.Unlock()
	// Now use IP range matches to find longest prefix match(es)
	if len(intervals) > 0 {
		customersFound := matchQueryPrefix(reverse(qname), intervals)
		found = len(customersFound)
		if found >= 1 {
			zone = customersFound[0].Zone
			name = customersFound[0].Name
		}
	}
	return zone, name, found
}

// Refresh selects all rows from customer database and updates internal
// interval tree.
func (c *CustomerDB) Refresh() error {
	atree := atree.New(1)
	mysql, err := sql.Open("mysql", DB_URL)
	if err != nil {
		return err
	}

	rows, err := mysql.Query(`
		SELECT
			id,
			name,
			TRIM(LEADING CHAR('\0') FROM ip_start) AS ip_start,
			TRIM(LEADING CHAR('\0') FROM ip_end) AS ip_end,
			zone
		FROM zones;`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		var name, zone string
		var ipStart, ipEnd []byte
		err := rows.Scan(&id, &name, &ipStart, &ipEnd, &zone)
		if err != nil {
			return err
		}
		cust := customer{
			Name: name,
			Zone: zone,
		}
		var ipStartInt, ipEndInt big.Int
		ipStartInt.SetBytes(ipStart)
		ipEndInt.SetBytes(ipEnd)
		ipRange := iprange.NewSingleDimensionInterval(ipStartInt, ipEndInt, id, cust)
		atree.Add(ipRange)
	}

	c.Lock()
	c.atree = &atree
	c.Unlock()
	return nil
}

// NewCustomerDB initializes the customer DB. Connects to mysql to fetch all
// data and build a radix tree.
func NewCustomerDB(path string) *CustomerDB {
	atree := atree.New(1)
	return &CustomerDB{
		new(sync.Mutex),
		path,
		&atree,
	}
}
