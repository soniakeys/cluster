// Author Sonia Keys 2012
// Public domain.

// K-means and K-means++ clustering for n-dimensional data
package cluster

import (
	"math/rand"
	"sort"
)

// KMeans, Lloyd's algorithm.
//
// Clusters points into k clusters where k = len(centers).  Initial values
// of centers are used as seed, or starting points for finding cluster centers.
//
// On return, centers will contain mean values of the discovered clusters.
//
// Also on return cNums will contain assigned cluster numbers for the input
// points and cCounts will contain the count of points in each cluster.
func KMeans(points []Point, centers []Point) (cNums, cCounts []int) {
	// working cluster number for each point
	cNums = make([]int, len(points))
	// initial assignment
	for i, p := range points {
		cNums[i], _ = p.NearestSqd(centers)
	}
	cCounts = make([]int, len(centers)) // size of each cluster
	for {
		// clear mean point values and cluster sizes
		for i, c := range centers {
			c.Clear()
			cCounts[i] = 0
		}
		// sum point values and counts
		for i, cx := range cNums {
			centers[cx].Add(points[i])
			cCounts[cx]++
		}
		for i := range centers { // compute means
			centers[i].Mul(1 / float64(cCounts[i]))
		}
		// make new assignments, count changes
		changes := false
		for i, p := range points {
			if cx, _ := p.NearestSqd(centers); cx != cNums[i] {
				changes = true
				cNums[i] = cx
			}
		}
		if !changes {
			return
		}
	}
}

// Kmpp, K-means++ clustering.
//
// Clusters points into k clusters.
//
// On return, centers will contain mean values of the discovered clusters,
// cNums will contain cluster numbers for points, cCounts will contain the
// count of points in each cluster.
//
// This is a wrapper for calling KMeans with the KmppSeeds initializer.
func Kmpp(points []Point, k int) (centers []Point, cNums, cCounts []int) {
	centers = KmppSeeds(points, k)
	cNums, cCounts = KMeans(points, centers)
	return
}

// KmppSeeds generates seed centers for KMeans
//
// KmppSeeds is the ++ part.  It picks the first seed randomly from points,
// then picks successive seeds randomly with probability proportional to the
// squared distance to the nearest seed.
//
// Randomness comes from math/rand default generator and is not seeded here.
func KmppSeeds(points []Point, k int) []Point {
	seeds := make([]Point, k) // return value
	// unselected points, initially all points
	up := make([]int, len(points))
	for i := range up {
		up[i] = i
	}
	// seed 0, selected randomly from all points
	s := rand.Intn(len(points))
	p := points[s]
	// d2 holds minimum sqd distances from points to any seed so far.
	d2 := make([]float64, len(points))
	for i, p2 := range points { // initialize here with sqd from seed 0
		d2[i] = p.Sqd(p2)
	}
	// upSum holds cumulative d2 distances for points in up.
	upSum := make([]float64, len(up))

	for sx := 0; ; {
		seeds[sx] = append(Point{}, points[s]...) // duplicate selected point
		sx++
		if sx == k {
			return seeds
		}
		// update d2
		if sx > 1 { // (first seed comes with d2 already done)
			for i, p2 := range points {
				if d := p.Sqd(p2); d < d2[i] {
					d2[i] = d
				}
			}
		}
		// remove selected point from up
		upLast := len(up) - 1
		up[s] = up[upLast]
		up = up[:upLast]
		// compute upSum
		sum := 0.
		for i, px := range up {
			sum += d2[px]
			upSum[i] = sum
		}
		// select next seed from up with probability proportional to d2
		s = up[sort.SearchFloat64s(upSum, rand.Float64()*sum)]
	}
}
