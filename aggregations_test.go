package main

import (
	"math"
	"testing"
)

var (
	noNulls   = floats(1.1, 2.2, 3.3)
	someNulls = floats(1.1, nil, 3.3)
	allNulls  = floats(nil, nil, nil)
)

type ex struct {
	// expecting an error
	err bool
	// expecting a result
	result float64
}

type example struct {
	// expected result when there are no nulls in the flattened list
	none ex
	// expected result when there are some nulls in the flattened list
	some ex
	// expected result when there are only nulls in the expected result
	all ex
}

// for each aggregation, we define 3 examples:
// none - what the aggregation should do when there are no nulls
// some - what the aggregation should do when there are some nulls
// all - what the aggregation should do when there are only nulls
// n.b., use coverage to ensure all aggregations are tested thoroughly
var examples = map[string]example{
	"sum":     {none: ex{false, 6.6}, some: ex{false, 4.4}, all: ex{true, 0.0}},
	"avg":     {none: ex{false, 6.6 / 3}, some: ex{false, 4.4 / 2}, all: ex{true, 0.0}},
	"min":     {none: ex{false, 1.1}, some: ex{false, 1.1}, all: ex{true, 0.0}},
	"max":     {none: ex{false, 3.3}, some: ex{false, 3.3}, all: ex{true, 0.0}},
	"median":  {none: ex{false, 2.2}, some: ex{false, 3.3}, all: ex{true, 0.0}},
	"95th":    {none: ex{false, 3.3}, some: ex{false, 3.3}, all: ex{true, 0.0}},
	"99th":    {none: ex{false, 3.3}, some: ex{false, 3.3}, all: ex{true, 0.0}},
	"999th":   {none: ex{false, 3.3}, some: ex{false, 3.3}, all: ex{true, 0.0}},
	"nullcnt": {none: ex{false, 0.0}, some: ex{false, 1.0}, all: ex{false, 3.0}},
	"nullpct": {none: ex{false, 0.0}, some: ex{false, 1.0 / 3.0}, all: ex{false, 1.0}},
}

func TestAggregations(t *testing.T) {
	for agg, ex := range examples {
		testSample(t, agg, "noNulls", noNulls, ex.none)
		testSample(t, agg, "someNulls", someNulls, ex.some)
		testSample(t, agg, "allNulls", allNulls, ex.all)
	}
}

func testSample(t *testing.T, agg string, name string, vs []*float64, e ex) {
	f := aggregations[agg]
	res, err := f(vs)
	if err != nil && !e.err {
		t.Errorf("unexpected error for %s on %s: %s", agg, name, err)
	} else if err == nil && e.err {
		t.Errorf("expected error for %s on %s, but got result %f", agg, name, res)
	} else if !eq(e.result, res) {
		t.Errorf("expected %f for %s on %s, but got %f", e.result, agg, name, res)
	}
}

func eq(a float64, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func floats(vs ...interface{}) []*float64 {
	res := make([]*float64, 0, len(vs))
	for _, v := range vs {
		if v == nil {
			res = append(res, nil)
		} else {
			f := v.(float64)
			res = append(res, &f)
		}
	}
	return res
}
