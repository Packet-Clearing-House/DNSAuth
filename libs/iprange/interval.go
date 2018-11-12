package iprange

import (
	"math/big"

	atree "github.com/Packet-Clearing-House/go-datastructures/augmentedtree"
)

// Dimension represents low and high of an interval at any dimension.
type Dimension struct {
	low, high big.Int
}

// IPInterval represents a set of dimensions and an identifier of an interval.
type IPInterval struct {
	dimensions []*Dimension
	id         uint64
	Value      interface{}
}

// NewSingleDimensionInterval is a convenience function that creates a single
// dimension interval for use with IP address ranges.
func NewSingleDimensionInterval(low, high big.Int, id uint64, value interface{}) *IPInterval {
	dimension := Dimension{
		low:  low,
		high: high,
	}
	return &IPInterval{
		dimensions: []*Dimension{&dimension},
		id:         id,
		Value:      value,
	}
}

// LowAtDimension returns an integer representing the lower bound at the
// requested dimension.
func (ip *IPInterval) LowAtDimension(dimension uint64) big.Int {
	return ip.dimensions[dimension-1].low
}

// HighAtDimension returns an integer representing the higher bound at the
// requested dimension.
func (ip *IPInterval) HighAtDimension(dimension uint64) big.Int {
	return ip.dimensions[dimension-1].high
}

// OverlapsAtDimension should return a bool indicating if the provided
// interval overlaps this interval at the dimension requested.
func (ip *IPInterval) OverlapsAtDimension(iv atree.Interval, dimension uint64) bool {
	ipHigh := ip.HighAtDimension(dimension)
	ipLow := ip.LowAtDimension(dimension)
	ivHigh := iv.HighAtDimension(dimension)
	ivLow := iv.LowAtDimension(dimension)
	return ipHigh.Cmp(&ivLow) > 0 && ipLow.Cmp(&ivHigh) < 0
}

// ID should be a unique ID representing this interval. This is used to
// identify which interval to delete from the tree if there are duplicates.
func (ip IPInterval) ID() uint64 {
	return ip.id
}
