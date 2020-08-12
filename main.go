package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/segfaultax/go-nagios"
	"github.com/segfaultax/go-nagios/util"
	"github.com/spf13/pflag"
)

var (
	showHelp     bool
	graphite     string
	warning      string
	critical     string
	username     string
	password     string
	metricName   string
	target       string
	from         string
	until        string
	aggregation  string
	timeout      int
	constantLine float64
	wrapLine     bool
)

type aggFunc func([]*float64) (float64, error)

var aggregations = map[string]aggFunc{
	"avg":     avgAgg,
	"sum":     sumAgg,
	"min":     minAgg,
	"max":     maxAgg,
	"median":  medianAgg,
	"95th":    q95Agg,
	"99th":    q99Agg,
	"999th":   q999Agg,
	"nullcnt": nullcntAgg,
	"nullpct": nullpctAgg,
}

const usage string = `usage: check-graphite [options]

The purpose of this tool is to check that the value given by a Graphite
query falls within certain warning and critical thresholds. Warning and
critical ranges can be provided in Nagios threshold format.

Example:

check-graphite -g localhost -m 'my.metric' -a sum -w 10 -c 100

Meaning: The sum of all non-null values returned by the Graphite query
'my.metric' is OK if less than or equal to 10, warning if greater than
10 but less than or equal to 100, critical if greater than 100. If it's
less than zero, it's critical.

Aggregations:

check-graphite supports the following aggregation functions:
* avg - mean average of all non-null values
* sum - sum of all non-null values
* min - minimum of all non-null values
* max - maximum of all non-null values
* median - median (50th percentile) of all non-null values
* 95th - 95th percentile of all non-null values
* 99th - 99th percentile of all non-null values
* 999th - 99.9th percentile of all non-null values
* nullcnt - count of null values
* nullpct - percentage of null values (nullcnt / total points)
`

func init() {
	pflag.BoolVarP(&showHelp, "help", "h", false, "show help")
	pflag.StringVarP(&graphite, "graphite", "g", "", "graphite host")

	pflag.StringVarP(&warning, "warning", "w", "", "warning range")
	pflag.StringVarP(&critical, "critical", "c", "", "critical range")

	pflag.StringVarP(&username, "username", "U", "", "username (HTTP Basic Auth)")
	pflag.StringVarP(&password, "password", "P", "", "password (HTTP Basic Auth)")

	pflag.StringVarP(&metricName, "name", "n", "metric", "Short, descriptive name for metric")
	pflag.StringVarP(&target, "target", "m", "", "Graphite query")

	pflag.StringVarP(&from, "from", "f", "1minute", "'from' value for query")
	pflag.StringVarP(&until, "until", "u", "", "'until' value for query")

	aggHelp := fmt.Sprintf("aggregation function, one of: %s", strings.Join(aggs(), ", "))

	pflag.StringVarP(&aggregation, "aggregation", "a", "avg", aggHelp)

	pflag.IntVarP(&timeout, "timeout", "t", 10, "Execution timeout")

	pflag.Float64VarP(&constantLine, "line", "l", 0.0, "the value used in constantLine(n) by --wrap")
	pflag.BoolVarP(&wrapLine, "wrap", "p", false, "wrap the query in a grouped constantLine(n) query")
}

func main() {
	pflag.Parse()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	err := checkRequiredOptions()
	if err != nil {
		printUsageErrorAndExit(3, err)
	}

	cli := newClient(graphite, username, password, timeout)

	check, err := nagios.NewRangeCheckParse(warning, critical)
	if err != nil {
		printUsageErrorAndExit(3, err)
	}
	defer check.Done()

	if wrapLine {
		target = fmt.Sprintf("group(%s, constantLine(%s))", target, util.PrettyFloat(constantLine, 6))
	}

	ms, err := cli.getMetrics(target, from, until)
	if err != nil {
		check.Unknown("failed to fetch metrics: %s", err)
		return
	}

	flat := flattenMetrics(ms)
	if len(flat) == 0 {
		check.Unknown("no metrics received from graphite")
		return
	}

	agg := aggregations[aggregation]
	val, err := agg(flat)
	if err != nil {
		check.Unknown(err.Error())
		return
	}

	check.CheckValue(val)
	check.AddPerfData(nagios.NewPerfData(aggregation, val, ""))
	check.SetMessage("%s (%s is %s)", metricName, aggregation, util.PrettyFloat(val, 6))
}

func checkRequiredOptions() error {
	_, ok := aggregations[aggregation]
	switch {
	case graphite == "":
		return fmt.Errorf("graphite is required")
	case target == "":
		return fmt.Errorf("target is required")
	case warning == "" && critical == "":
		return fmt.Errorf("must supply at least one of -w or -c")
	case !ok:
		return fmt.Errorf("aggregation must be one of: %s", strings.Join(aggs(), ", "))
	}
	return nil
}

func printUsageErrorAndExit(code int, err error) {
	fmt.Printf("execution failed: %s\n", err)
	printUsage()
	os.Exit(code)
}

func printUsage() {
	fmt.Println(usage)
	pflag.PrintDefaults()
}

func aggs() []string {
	var aggs []string
	for agg := range aggregations {
		aggs = append(aggs, agg)
	}
	return aggs
}
