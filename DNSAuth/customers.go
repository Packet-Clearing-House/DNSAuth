package main

import (
	"bytes"
	"database/sql"
	"net"
	"sync"

	radix "github.com/armon/go-radix"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DB_URL = "root:pass@(127.0.0.1)/customers"

type customer struct {
	Name    string
	IPStart []byte
	IPEnd   []byte
}

// Function that reverse a word (test.com -> moc.tset)
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

type CustomerDB struct {
	*sync.Mutex
	dburl string
	tree  *radix.Tree
}

// Resolve the customer name from DNS qname
// Returns Unknown if not found
func (c *CustomerDB) Resolve(qname string, ip net.IP) (string, string) {

	name := "Unknown"
	zone := "Unknown"
	c.Lock()
	zone, value, found := c.tree.LongestPrefix(reverse(qname))
	c.Unlock()
	if found {
		cust := value.(customer)
		if bytes.Compare(ip, cust.IPStart) >= 0 && bytes.Compare(ip, cust.IPEnd) <= 0 {
			name = cust.Name
		}
		name = cust.Name
	}
	return reverse(zone), name
}

func (c *CustomerDB) Refresh() error {

	tree := radix.New()
	mysql, err := sql.Open("mysql", DB_URL)
	if err != nil {
		return err
	}

	rows, err := mysql.Query("SELECT name, ip_start, ip_end, zone FROM zones;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, zone string
		var ipStart, ipEnd []byte
		err := rows.Scan(&name, &ipStart, &ipEnd, &zone)
		if err != nil {
			return err
		}
		cust := customer{
			Name:    name,
			IPStart: ipStart,
			IPEnd:   ipEnd,
		}
		tree.Insert(reverse(zone), cust)
	}

	c.Lock()
	c.tree = tree
	c.Unlock()
	return nil
}

// Init the customer DB. Connects to mysql to fetch all data and build a radix tree
func NewCustomerDB(path string) *CustomerDB {

	return &CustomerDB{
		new(sync.Mutex),
		path,
		radix.New(),
	}
}
