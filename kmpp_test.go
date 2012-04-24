package kmpp

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

// Method:  Generate a bunch of random points normally distributed around
// each of two points, o1 and o2.  Have Kmpp find two clusters, c1 and c2.
// compute d1 = distance(o1, c1) + distance(o2, c2)
// compute d2 = distance(o1, c2) + distance(o2, c1)
// Then one of d1, d2 should be near zero and one of them should be near zero
// and the other should be near 2*distance(o1, o2).
func TestKmpp(t *testing.T) {
	o1 := Point{100, 140, 80}
	o2 := Point{200, 160, 120}
	o := []Point{o1, o2}

	// dimensionality of points, = 3.
	dim := len(o1)
	// the k of k-means, the number of clusters to partition into, = 2.
	k := len(o)

	const nPointsPerCluster = 1000
	const stdv = 80.

	rand.Seed(time.Now().UnixNano())
	data := make([]CPoint, k*nPointsPerCluster)
	p := 0
	for n := range data {
		data[n].Point = make(Point, dim)
		for i, x := range o[p] {
			data[n].Point[i] = rand.NormFloat64()*stdv + x
		}
		data[n].C = p
		p = 1 - p
	}

	Kmpp(k, data)

	// clustering done, compute statistics.  first accumulate by cluster:
	cCent := make([]Point, k)
	for i := range cCent {
		cCent[i] = make(Point, dim)
	}
	cLen := make([]int, k)
	for _, p := range data {
		cLen[p.C]++             // count by cluster
		cCent[p.C].Add(p.Point) // sum by cluster
	}
	inv := make([]float64, k)
	for i, iLen := range cLen {
		inv[i] = 1 / float64(iLen) // compute 1/count by cluster
		cCent[i].Mul(inv[i])       // compute mean by cluster
	}
	c1 := cCent[0]
	c2 := cCent[1]

	// log a bunch of information
	t.Log("Data set origins:")
	t.Logf("%5.1f", o1)
	t.Logf("%5.1f", o2)

	t.Log("Cluster centroids, mean distance from centroid,")
	t.Log("number of points in cluster:")
	t.Logf(" %*s  distance  points", dim*6, "centroid")
	dist := make([]float64, k)
	for _, p := range data {
		dist[p.C] += math.Sqrt(p.Sqd(cCent[p.C]))
	}
	t.Logf("%5.1f  %8.1f  %6d\n", c1, dist[0]*inv[0], cLen[0])
	t.Logf("%5.1f  %8.1f  %6d\n", c2, dist[1]*inv[1], cLen[1])

	// compute d1, d2 for test
	d1 := math.Sqrt(o1.Sqd(c1)) + math.Sqrt(o2.Sqd(c2))
	d2 := math.Sqrt(o1.Sqd(c2)) + math.Sqrt(o2.Sqd(c1))
	if d1 > stdv && d2 > stdv {
		t.Log("cluster centers far from original clusters")
		t.Log("d1:", d1, "d2:", d2)
	}
}
