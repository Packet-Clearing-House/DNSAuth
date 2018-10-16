package main

import (
	"database/sql"
	"sync"

	radix "github.com/armon/go-radix"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var DB_URL = "root:pass@(127.0.0.1)/customers"

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
func (c *CustomerDB) Resolve(qname string) (string, string) {

	name := "Unknown"
	zone := "Unknown"
	c.Lock()
	zone, value, found := c.tree.LongestPrefix(reverse(qname))
	c.Unlock()
	if found {
		name = value.(string)
	}
	return reverse(zone), name
}

func (c *CustomerDB) Refresh() error {

	tree := radix.New()
	mysql, err := sql.Open("mysql", DB_URL)
	if err != nil {
		return err
	}

	rows, err := mysql.Query("SELECT group_name, zone FROM zones;")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, zone string
		err := rows.Scan(&name, &zone)
		if err != nil {
			return err
		}
		tree.Insert(reverse(zone), name)
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
