// Copyright 2012 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

// K-means and K-means++ clustering for n-dimensional data
package kmpp

import (
	"math"
	"math/rand"
)

// n-dimensional point
type Point []float64

// Sqd, square of distance between Points
func (p1 Point) Sqd(p2 Point) (ssq float64) {
	for i, x1 := range p1 {
		d := x1 - p2[i]
		ssq += d * d
	}
	return
}

// Add, element-wise += on a Point.
func (p1 Point) Add(p2 Point) {
	for i, x2 := range p2 {
		p1[i] += x2
	}
}

// Mul, scalar multipy on a Point.
func (p Point) Mul(s float64) {
	for i := range p {
		p[i] *= s
	}
}

// CPoint, clustered point, a point associated with a cluster.
type CPoint struct {
	C int // cluster number
	Point
}

// Kmpp, K-means++.
func Kmpp(k int, data []CPoint) {
	KMeans(data, kmppSeeds(k, data))
}

// kmppSeeds is the ++ part.
// It generates the initial means for the k-means algorithm.
func kmppSeeds(k int, data []CPoint) []Point {
	s := make([]Point, k)
	s[0] = append(Point{}, data[rand.Intn(len(data))].Point...)
	d2 := make([]float64, len(data))
	for i := 1; i < k; i++ {
		var sum float64
		for j, p := range data {
			_, dMin := nearest(p, s[:i])
			d2[j] = dMin * dMin
			sum += d2[j]
		}
		target := rand.Float64() * sum
		j := 0
		for sum = d2[0]; sum < target; sum += d2[j] {
			j++
		}
		s[i] = append(Point{}, data[j].Point...)
	}
	return s
}

// nearest finds the nearest mean to a given point.
// return values are the index of the nearest mean, and the distance from
// the point to the mean.
func nearest(p CPoint, mean []Point) (int, float64) {
	iMin := 0
	sqdMin := p.Sqd(mean[0])
	for i := 1; i < len(mean); i++ {
		sqd := p.Sqd(mean[i])
		if sqd < sqdMin {
			sqdMin = sqd
			iMin = i
		}
	}
	return iMin, math.Sqrt(sqdMin)
}

// KMeans, Lloyd's algorithm.
func KMeans(data []CPoint, mean []Point) {
	// initial assignment
	for i, p := range data {
		cMin, _ := nearest(p, mean)
		data[i].C = cMin
	}
	mLen := make([]int, len(mean))
	for n := len(data[0].Point); ; {
		// update means
		for i := range mean {
			mean[i] = make(Point, n)
			mLen[i] = 0
		}
		for _, p := range data {
			mean[p.C].Add(p.Point)
			mLen[p.C]++
		}
		for i := range mean {
			mean[i].Mul(1 / float64(mLen[i]))
		}
		// make new assignments, count changes
		var changes int
		for i, p := range data {
			if cMin, _ := nearest(p, mean); cMin != p.C {
				changes++
				data[i].C = cMin
			}
		}
		if changes == 0 {
			return
		}
	}
}
