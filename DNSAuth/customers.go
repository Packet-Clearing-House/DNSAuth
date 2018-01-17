package main

import (
	"database/sql"
	"log"
	"github.com/asergeyev/nradix"
	_"github.com/lib/pq"
)

type Customer struct {
	Prefix string
	Name string
	PrefixMonit bool
	ASNMonit bool
}

var DB_URL = "postgres://user@127.0.0.1/pipeline?sslmode=disable"

func getCustomerTree() (*nradix.Tree, error) {

	tree := nradix.NewTree(0)

	db, err := sql.Open("postgres", DB_URL)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT ip::cidr, name, asn, prefix FROM ns_customers;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	for rows.Next() {
		c := Customer{}
		err := rows.Scan(&c.Prefix, &c.Name, &c.ASNMonit, &c.PrefixMonit)
		if err != nil {
			log.Fatal(err)
		}
		err = tree.AddCIDR(c.Prefix, &c)
		if err != nil {
			log.Println(err)
		}
	}
	return tree, nil
}