package main

import (
	"database/sql"
	"log"
	_ "github.com/go-sql-driver/mysql"
	radix "github.com/armon/go-radix"
	_"github.com/lib/pq"
)

var DB_URL = "root:pass@(127.0.0.1)/customerdb"

// Function that reverse a word (test.com -> moc.tset)
func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

type CustomerDB struct {
	tree *radix.Tree
}

// Resolve the customer name from DNS qname 
// Returns Unknown if not found
func (c *CustomerDB) Resolve(qname string) string {
	name := "Unknown"
	_, value, found := c.tree.LongestPrefix(reverse(qname))
	if found {
		name = value.(string)
	}
	return name
}

// Init the customer DB. Connects to mysql to fetch all data and build a radix tree
func InitCustomerDB(path string) (*CustomerDB, error) {
	db := &CustomerDB{
		radix.New(),
	}

	mysql, err := sql.Open("mysql", DB_URL)
	if err != nil {
		return nil, err
	}

	rows, err := mysql.Query("SELECT group_name, zone FROM zones;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	for rows.Next() {
		var name, zone string
		err := rows.Scan(&name, &zone)
		if err != nil {
			log.Fatal(err)
		}
		db.tree.Insert(reverse(zone), name)
	}
	
	return db, nil
}