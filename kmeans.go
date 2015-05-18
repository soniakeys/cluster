// Author Sonia Keys 2012
// Public domain.

package cluster

// K-means and K-means++ clustering for n-dimensional data

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
// Distortion is the squared error distortion, a measure of how well the
// data clustered.
func KMeans(points, centers []Point) (cNums, cCounts []int, distortion float64) {
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
		distortion = 0
		for i, p := range points {
			cx, sqd := p.NearestSqd(centers)
			distortion += sqd
			if cx != cNums[i] {
				changes = true
				cNums[i] = cx
			}
		}
		if !changes {
			distortion /= float64(len(points))
			return
		}
	}
}

// KMPP, K-means++ clustering.
//
// Clusters points into k clusters.
//
// On return, centers will contain mean values of the discovered clusters,
// cNums will contain cluster numbers for points, cCounts will contain the
// count of points in each cluster.
//
// This is a wrapper for calling KMeans with the KMSeedPP initializer.
func KMPP(points []Point, k int) (centers []Point, cNums, cCounts []int, distortion float64) {
	centers = KMSeedPP(points, k)
	cNums, cCounts, distortion = KMeans(points, centers)
	return
}

// KMSeedPP generates seed centers for KMeans using the KMeans++ initializer.
//
// KMSeedPP is the ++ part.  It picks the first seed randomly from points,
// then picks successive seeds randomly with probability proportional to the
// squared distance to the nearest seed.
//
// Randomness comes from math/rand default generator and is not seeded here.
//
// Returned seeds are copies of the selected points.
func KMSeedPP(points []Point, k int) []Point {
	seeds := make([]Point, k)           // return value
	p := points[rand.Intn(len(points))] // select first seed randomly
	d2 := make([]float64, len(points))  // minimum sqd to any seed
	for i, p2 := range points {         // initialize d2
		d2[i] = p.Sqd(p2)
	}
	dSum := make([]float64, len(points)) // cumulative d2 distances
	for sx := 0; ; {
		seeds[sx] = append(Point{}, p...) // duplicate selected point
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
		// compute dSum
		sum := 0.
		for i, d := range d2 {
			sum += d
			dSum[i] = sum
		}
		// select next seed from up with probability proportional to d2
		p = points[sort.SearchFloat64s(dSum, rand.Float64()*sum)]
	}
}

// KMSeedRandom selects and copies k distinct points from the points argument.
//
// Randomness comes from math/rand default generator and is not seeded here.
//
// The function panics if there are not k distinct points.
func KMSeedRandom(points []Point, k int) []Point {
	seeds := make([]Point, k)
	for i, s := range rand.Perm(len(points))[:k] {
		seeds[i] = append(Point{}, points[s]...)
	}
	return seeds
}

// KMSeedFirst simply copies the first k points.
//
// The function panics if len(points) < k.
func KMSeedFirst(points []Point, k int) []Point {
	seeds := make([]Point, k)
	for i, p := range points[:k] {
		seeds[i] = append(Point{}, p...)
	}
	return seeds
}
