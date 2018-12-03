package main

import (
	"math/big"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Packet-Clearing-House/DNSAuth/libs/iprange"
	atree "github.com/Packet-Clearing-House/go-datastructures/augmentedtree"
)

type customerRow struct {
	ipStart  []byte
	ipEnd    []byte
	customer customer
}

var cdb *CustomerDB

func init() {
	atree := atree.New(1)
	customerRows := customerData()
	for i, c := range customerRows {
		var ipStartInt, ipEndInt big.Int
		ipStartInt.SetBytes(c.ipStart)
		ipEndInt.SetBytes(c.ipEnd)
		id := uint64(i)
		ipRange := iprange.NewSingleDimensionInterval(ipStartInt, ipEndInt, id, c.customer)
		atree.Add(ipRange)
	}
	cdb = &CustomerDB{
		new(sync.Mutex),
		"",
		&atree,
	}
}

// customerData returns customer rows to use as fixtures for tests
func customerData() []customerRow {
	customerRows := []customerRow{
		customerRow{
			ipBytes(net.ParseIP("100.100.100.0")),
			ipBytes(net.ParseIP("100.100.100.100")),
			customer{"customer-one", "one.com"},
		},
		customerRow{
			ipBytes(net.ParseIP("1.2.3.4")),
			ipBytes(net.ParseIP("1.2.3.4")),
			customer{"customer-two", "two.com"},
		},
		customerRow{
			ipBytes(net.ParseIP("199.0.0.0")),
			ipBytes(net.ParseIP("199.0.0.40")),
			customer{"overlap-one", "over1.com"},
		},
		customerRow{
			ipBytes(net.ParseIP("199.0.0.30")),
			ipBytes(net.ParseIP("199.0.0.50")),
			customer{"overlap-two", "over2.com"},
		},
		customerRow{
			ipBytes(net.ParseIP("fdfe::5a55:caff:fefa:9000")),
			ipBytes(net.ParseIP("fdfe::5a55:caff:fefa:9089")),
			customer{"foobar6", "foobar6.com"},
		},
	}
	return customerRows
}

func TestMatches(t *testing.T) {
	// Match IPv4 in single range
	foundZone, foundName, foundNum := cdb.Resolve("foo.one.com", net.ParseIP("100.100.100.10"))
	assert.Equal(t, "one.com", foundZone)
	assert.Equal(t, "customer-one", foundName)
	assert.Equal(t, 1, foundNum)
	// Match IPv4 as a single IP
	foundZone, foundName, foundNum = cdb.Resolve("foo.two.com", net.ParseIP("1.2.3.4"))
	assert.Equal(t, "two.com", foundZone)
	assert.Equal(t, "customer-two", foundName)
	assert.Equal(t, 1, foundNum)
	// Match single one out of overlapping ranges based on prefix
	foundZone, foundName, foundNum = cdb.Resolve("foo.over1.com", net.ParseIP("199.0.0.35"))
	assert.Equal(t, "over1.com", foundZone)
	assert.Equal(t, "overlap-one", foundName)
	assert.Equal(t, 1, foundNum)
	// Match single one out of overlapping ranges based on prefix
	foundZone, foundName, foundNum = cdb.Resolve("foo.over2.com", net.ParseIP("199.0.0.35"))
	assert.Equal(t, "over2.com", foundZone)
	assert.Equal(t, "overlap-two", foundName)
	assert.Equal(t, 1, foundNum)
	// Match IPv6 in single range
	foundZone, foundName, foundNum = cdb.Resolve("baz.foobar6.com", net.ParseIP("fdfe::5a55:caff:fefa:9010"))
	assert.Equal(t, "foobar6.com", foundZone)
	assert.Equal(t, "foobar6", foundName)
	assert.Equal(t, 1, foundNum)
	// Unknown
	foundZone, foundName, foundNum = cdb.Resolve("foobarbaz.com", net.ParseIP("1.1.1.1"))
	assert.Equal(t, "Unknown", foundZone)
	assert.Equal(t, "Unknown", foundName)
	assert.Equal(t, 0, foundNum)
}
