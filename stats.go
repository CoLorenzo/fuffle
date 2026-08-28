package main

import (
	"math"
	"sort"
)

func Mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func StdDev(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := Mean(data)
	sum := 0.0
	for _, v := range data {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(data)))
}

func Median(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

func Mode(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	bucketSize := 100.0
	buckets := make(map[int64]int)
	for _, v := range data {
		bucket := int64(v / bucketSize)
		buckets[bucket]++
	}
	maxCount := 0
	var bestBucket int64
	for bucket, count := range buckets {
		if count > maxCount {
			maxCount = count
			bestBucket = bucket
		}
	}
	return float64(bestBucket)*bucketSize + bucketSize/2
}
