package main

import (
	"fmt"
	"math"
	"sort"
)

func sumAgg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	sum := 0.0
	for _, v := range vs {
		sum += v
	}

	return sum, nil
}

func avgAgg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	sum := 0.0
	for _, v := range vs {
		sum += v
	}

	return sum / float64(len(vs)), nil
}

func maxAgg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	max := math.Inf(-1)
	for _, v := range vs {
		if v > max {
			max = v
		}
	}

	return max, nil
}

func minAgg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	min := math.Inf(1)
	for _, v := range vs {
		if v < min {
			min = v
		}
	}

	return min, nil
}

func medianAgg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	return quantile(vs, 0.5), nil
}

func q95Agg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	return quantile(vs, 0.95), nil
}

func q99Agg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	return quantile(vs, 0.99), nil
}

func q999Agg(metrics []*float64) (float64, error) {
	vs := removeNulls(metrics)
	if len(vs) == 0 {
		return 0, fmt.Errorf("only null values returned")
	}

	return quantile(vs, 0.999), nil
}

func nullcntAgg(metrics []*float64) (float64, error) {
	cnt := 0
	for _, v := range metrics {
		if v == nil {
			cnt++
		}
	}

	return float64(cnt), nil
}

func nullpctAgg(metrics []*float64) (float64, error) {
	cnt := 0
	for _, v := range metrics {
		if v == nil {
			cnt++
		}
	}

	return float64(cnt) / float64(len(metrics)), nil
}

func flattenMetrics(ms metrics) []*float64 {
	size := 0
	for _, m := range ms {
		size += len(m.Datapoints)
	}

	vals := make([]*float64, 0, size)
	for _, m := range ms {
		for _, d := range m.Datapoints {
			vals = append(vals, d.Value)
		}
	}

	return vals
}

func removeNulls(vs []*float64) []float64 {
	xs := make([]float64, 0, len(vs))
	for _, v := range vs {
		if v != nil {
			xs = append(xs, *v)
		}
	}
	return xs
}

func quantile(vs []float64, q float64) float64 {
	sort.Float64s(vs)
	return vs[int(float64(len(vs))*q)]
}
