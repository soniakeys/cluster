// Author Sonia Keys 2012
// Public domain.

package cluster

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/soniakeys/cluster"
)

// Method:  Generate a bunch of random points normally distributed around
// each of two points, o1 and o2.  Have KMPP find two clusters, c1 and c2.
// compute d1 = distance(o1, c1) + distance(o2, c2)
// compute d2 = distance(o1, c2) + distance(o2, c1)
// Then one of d1, d2 should be near zero and one of them should be near zero
// and the other should be near 2*distance(o1, o2).
func TestKMPP(t *testing.T) {
	o1 := cluster.Point{100, 140, 80}
	o2 := cluster.Point{200, 160, 120}
	o := []cluster.Point{o1, o2}

	// dimensionality of points, = 3.
	dim := len(o1)
	// the k of k-means, the number of clusters to partition into, = 2.
	k := len(o)
	const stdv = 80.
	const nPointsPerCluster = 1000

	rand.Seed(time.Now().UnixNano())
	data := make([]cluster.Point, k*nPointsPerCluster)
	p := 0
	for n := range data {
		data[n] = make(cluster.Point, dim)
		for i, x := range o[p] {
			data[n][i] = rand.NormFloat64()*stdv + x
		}
		p = 1 - p
	}
	cCent, _, cLen, _ := cluster.KMPP(data, k)
	c1 := cCent[0]
	c2 := cCent[1]

	// log a bunch of information
	t.Log("Data set origins:")
	t.Logf("%5.1f", o1)
	t.Logf("%5.1f", o2)

	t.Log("Cluster centroids, number of points in cluster:")
	t.Logf(" %*s  points", dim*6, "centroid")
	t.Logf("%5.1f  %6d\n", c1, cLen[0])
	t.Logf("%5.1f  %6d\n", c2, cLen[1])

	// compute d1, d2 for test
	d1 := math.Sqrt(o1.Sqd(c1)) + math.Sqrt(o2.Sqd(c2))
	d2 := math.Sqrt(o1.Sqd(c2)) + math.Sqrt(o2.Sqd(c1))
	if d1 > stdv && d2 > stdv {
		t.Log("cluster centers far from original clusters")
		t.Log("d1:", d1, "d2:", d2)
	}
}
