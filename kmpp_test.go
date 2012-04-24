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
	o1 := R2{100, 80}
	o2 := R2{200, 120}
	o := []R2{o1, o2}

	const nPointsPerCluster = 1000
	const stdv = 50.

	rand.Seed(time.Now().UnixNano())
	data := make([]R2c, 2*nPointsPerCluster)
	p := 0
	for n := range data {
		data[n].X = rand.NormFloat64()*stdv + o[p].X
		data[n].Y = rand.NormFloat64()*stdv + o[p].Y
		data[n].C = p
		p = 1 - p
	}

	// the k of k-means, the number of clusters to partition into, = 2.
	k := len(o)
	Kmpp(k, data)

	// clustering done, compute statistics.  first accumulate by cluster:
	c := make([]R2, k)
	cLen := make([]int, k)
	for _, p := range data {
		cLen[p.C]++     // count by cluster
		c[p.C].X += p.X // sum by cluster
		c[p.C].Y += p.Y
	}
	inv := make([]float64, k)
	for i, iLen := range cLen {
		inv[i] = 1 / float64(iLen) // compute 1/count by cluster
		c[i].X *= inv[i]           // compute mean by cluster
		c[i].Y *= inv[i]
	}
	c1 := c[0]
	c2 := c[1]

	// log a bunch of information
	t.Log("Data set origins:")
	t.Log("    x      y")
	t.Logf("%5.1f  %5.1f\n", o1.X, o1.Y)
	t.Logf("%5.1f  %5.1f\n", o2.X, o2.Y)

	t.Log("Cluster centroids, mean distance from centroid, number of points:")
	t.Log("    x      y  distance  points")
	dist := make([]float64, k)
	for _, p := range data {
		dist[p.C] += math.Hypot(p.X-c[p.C].X, p.Y-c[p.C].Y)
	}
	t.Logf("%5.1f  %5.1f  %8.1f  %6d\n", c1.X, c1.Y, dist[0]*inv[0], cLen[0])
	t.Logf("%5.1f  %5.1f  %8.1f  %6d\n", c2.X, c2.Y, dist[1]*inv[1], cLen[1])

	// compute d1, d2 for test
	d1 := math.Hypot(o1.X-c1.X, o1.Y-c1.Y) + math.Hypot(o2.X-c2.X, o2.Y-c2.Y)
	d2 := math.Hypot(o1.X-c2.X, o1.Y-c2.Y) + math.Hypot(o2.X-c1.X, o2.Y-c1.Y)
	if d1 > stdv && d2 > stdv {
		t.Log("cluster centers far from original clusters")
		t.Log("d1:", d1, "d2:", d2)
	}
}
