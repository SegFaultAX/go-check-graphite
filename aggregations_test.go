package main

import (
	"fmt"
	"math"
	"testing"
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

var samples = map[string][]*float64{
  "noNulls":    floats(1.1, 2.2, 3.3),
  "someNulls":  floats(1.1, nil, 3.3),
  "allNulls":   floats(nil, nil, nil),
}

// for each aggregation, we define 3 examples:
// none - what the aggregation should do when there are no nulls
// some - what the aggregation should do when there are some nulls
// all - what the aggregation should do when there are only nulls
// n.b., use coverage to ensure all aggregations are tested thoroughly
var examples = []struct{
  agg string; sample string; expected float64 }{
	{"avg",      "noNulls",    6.6 / 3,    },
	{"avg",      "someNulls",  4.4 / 2,    },
	{"max",      "noNulls",    3.3,        },
	{"max",      "someNulls",  3.3,        },
	{"median",   "noNulls",    2.2,        },
	{"median",   "someNulls",  3.3,        },
	{"min",      "noNulls",    1.1,        },
	{"min",      "someNulls",  1.1,        },
	{"nullcnt",  "allNulls",   3.0,        },
	{"nullcnt",  "noNulls",    0.0,        },
	{"nullcnt",  "someNulls",  1.0,        },
	{"nullpct",  "allNulls",   1.0,        },
	{"nullpct",  "noNulls",    0.0,        },
	{"nullpct",  "someNulls",  1.0 / 3.0,  },
	{"sum",      "noNulls",    6.6,        },
	{"sum",      "someNulls",  4.4,        },
	{"95th",     "noNulls",    3.3,        },
	{"95th",     "someNulls",  3.3,        },
	{"999th",    "noNulls",    3.3,        },
	{"999th",    "someNulls",  3.3,        },
	{"99th",     "noNulls",    3.3,        },
	{"99th",     "someNulls",  3.3,        },
}


var errExamples = []struct{
  agg string; sample string }{
	{"avg",     "allNulls",  },
	{"max",     "allNulls",  },
	{"median",  "allNulls",  },
	{"min",     "allNulls",  },
	{"sum",     "allNulls",  },
	{"95th",    "allNulls",  },
	{"999th",   "allNulls",  },
	{"99th",    "allNulls",  },
}

func TestAggregationsWithoutErrors(t *testing.T) {
	for _, ex := range examples {
		testNoError(t, ex.agg, ex.sample, samples[ex.sample], ex.expected)
	}
}

func TestAggregationsWithErrors(t *testing.T) {
	for _, ex := range errExamples {
		testExpectError(t, ex.agg, ex.sample, samples[ex.sample])
	}
}

func testNoError(t *testing.T, agg string, name string, vs []*float64, expected float64) {
	t.Run(fmt.Sprintf("%s - %s expecting no error", agg, name), func(t *testing.T) {
		f := aggregations[agg]
		actual, err := f(vs)
		if err != nil {
			t.Errorf("unexpected error for %s on %s: %s", agg, name, err)
		}
    if !eq(expected, actual) {
			t.Errorf("expected %f for %s on %s, but got %f", expected, agg, name, actual)
		}
	})
}

func testExpectError(t *testing.T, agg string, name string, vs []*float64) {
	t.Run(fmt.Sprintf("%s - %s, expecting error", agg, name), func(t *testing.T) {
    expected := 0.0
		f := aggregations[agg]
		actual, err := f(vs)
    if err == nil {
			t.Errorf("expected error for %s on %s, but got result %f", agg, name, actual)
		}
    if !eq(expected, actual) {
			t.Errorf("expected %f for %s on %s, but got %f", expected, agg, name, actual)
		}
	})
}

func eq(a float64, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func floats(vs ...interface{}) []*float64 {
	actual := make([]*float64, 0, len(vs))
	for _, v := range vs {
		if v == nil {
			actual = append(actual, nil)
		} else {
			f := v.(float64)
			actual = append(actual, &f)
		}
	}
	return actual
}
