// Public domain.

package cluster

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"math/rand"

	"github.com/soniakeys/bits"
	"github.com/soniakeys/graph"
)

// DistanceMatrix holds a distance matrix for cluster analyses.
//
// See DistanceMatrix.Valid for typical restrictions.
type DistanceMatrix [][]float64

// String returns a string representation of a DistanceMatrix.
//
// The function is simplistic, useful mostly for debugging.  For more attractive
// output or more output options, consider a more general matix library such
// as github.com/gonum/matrix/mat64 or github.com/skelterjohn/go.matrix.
func (d DistanceMatrix) String() string {
	if len(d) == 0 {
		return ""
	}
	var b bytes.Buffer
	fmt.Fprint(&b, d[0])
	for _, di := range d[1:] {
		fmt.Fprintf(&b, "\n%v", di)
	}
	return b.String()
}

// Clone allocates and copies a DistanceMatrix
func (d DistanceMatrix) Clone() DistanceMatrix {
	dc := make(DistanceMatrix, len(d))
	for i, di := range d {
		dc[i] = append([]float64{}, di...)
	}
	return dc
}

// NewEuclideanDist constructs an n×n distance matrix where n is len(exp)
// based on Euclidean distance between points.
func NewEuclideanDist(exp []Point) DistanceMatrix {
	dist := make(DistanceMatrix, len(exp))
	for i := range dist {
		di := make([]float64, len(exp))
		for j := 0; j < i; j++ {
			d := math.Sqrt(exp[i].Sqd(exp[j]))
			di[j] = d
			dist[j][i] = d
		}
		dist[i] = di
	}
	return dist
}

// Square tests if a DistanceMatrix is square.
func (d DistanceMatrix) Square() bool {
	for _, di := range d {
		if len(di) != len(d) {
			return false
		}
	}
	return true
}

// NonNegative tests that a DistanceMatrix has no negative elements.
//
// The test is NaN weak -- a NaN element does not count as negative and so
// will not cause the function to return false.
func (d DistanceMatrix) NonNegative() bool {
	for _, di := range d {
		for _, dij := range di {
			if dij < 0 {
				return false
			}
		}
	}
	return true
}

// Symmetric tests if off-diagonal elements of a DistanceMatrix are symmetric.
func (d DistanceMatrix) Symmetric() bool {
	for i, di := range d {
		for j, dij := range di[:i] {
			if !(dij == d[j][i]) { // reversed test catches NaNs too.
				return false
			}
		}
	}
	return true
}

// ZeroDiagonal tests that all diagonal elements are zero.
func (d DistanceMatrix) ZeroDiagonal() bool {
	for i, di := range d {
		if !(di[i] == 0) {
			return false
		}
	}
	return true
}

// TriangleInequality tests that there are no violations of the triangle
// inequality for distance matrices.
//
// That is, it tests that d[i][j] + d[k][j] < d[i][k] for all i, j, k.
//
// The test is NaN weak -- the presence of a NaN does not cause the function
// to return false.
func (d DistanceMatrix) TriangleInequality() (ok bool, i, j, k int) {
	for i, di := range d {
		for k, dk := range d[:i] {
			dik := di[k]
			for j, dij := range di {
				if dij+dk[j] < dik {
					return false, i, j, k
				}
			}
		}
	}
	return true, 0, 0, 0
}

// Valid validates a DistanceMarix as Euclidean.
//
// Conditions are:
//
//   * square:  len(d[i]) == len(d)
//   * symmetric:  d[i][j] == d[j][i]
//   * non-negative:  d[i][j] >= 0
//   * zero diagonal:  d[i][i] == 0
//   * triangle inequality:  d[i][j] + d[j][k] <= d[i][k]
//
// Valid returns nil if all conditions are met, otherwise an error citing
// a condition not met.
func (d DistanceMatrix) Validate() error {
	if !d.Square() {
		return errors.New("not square")
	}
	if !d.NonNegative() {
		return errors.New("negative element")
	}
	if !d.Symmetric() {
		return errors.New("not symmetric")
	}
	if !d.ZeroDiagonal() {
		return errors.New("non-zero diagonal")
	}
	if ok, i, j, k := d.TriangleInequality(); !ok {
		return fmt.Errorf("triangle inequality not satisfied: "+
			"d[%d][%d] + d[%d][%d] < d[%d][%d]", i, j, j, k, i, k)
	}
	return nil
}

// Additive tests if DistanceMatrix d is additive.
//
// Additive tests the four-point condition for all combinations of points.
// If a test fails, it returns ok = false and the four failing points.
func (d DistanceMatrix) Additive() (ok bool, i, j, k, l int) {
	for i, di := range d {
		for j, dj := range d[:i] {
			dij := di[j]
			for k, dk := range d[:j] {
				dik := di[k]
				djk := dj[k]
				for l, dil := range di {
					s1 := dij + dk[l]
					s2 := dik + dj[l]
					s3 := dil + djk
					switch { // swap so s1 is the value !=
					case s1 == s2:
						s1, s3 = s3, s1
					case s1 == s3:
						s1, s2 = s2, s1
					}
					if s1 > s2 {
						return false, i, j, k, l
					}
				}
			}
		}
	}
	ok = true
	return
}

// limbWeight finds the weight of the edge to a leaf (a limb) in a phylogenic
// tree corresponding to DistanceMatrix d.
//
// Argument j is an index of d representing a leaf node.
//
// Returned is the weight wt of an edge that would connect j to a phylogenic
// tree, and two other leaf indexes, i and k, such that the edge connecting j
// to the tree would connect directly to the path between i and k.
func (d DistanceMatrix) limbWeight(j int) (wt float64, i, k int) {
	// algorithm: pick k for convenience in iteration, iterate i over all
	// other indexes, accumulating the value of i that minimizes wt.
	var dj0, dj1, dk0, dk1 []float64
	if j == 0 {
		k = 1
		dj1 = d[j][2:]
		dk1 = d[k][2:]
	} else {
		k = j - 1
		dj0 = d[j][:k]
		dk0 = d[k][:k]
		dj1 = d[j][j+1:]
		dk1 = d[k][j+1:]
	}
	djk := d[j][k]
	var iMin2 int      // index i minimizing wtMin2
	var wtMin2 float64 // accumulated sum, twice min, return wtMin2/2
	if len(dj0) > 0 {
		iMin2 = len(dj0) - 1
		dj0 = dj0[:iMin2]
		dk0 = dk0[:iMin2]
	} else {
		iMin2 = 2
		dj1 = d[j][j+2:]
		dk1 = d[k][j+2:]
	}
	wtMin2 = d[j][iMin2] + djk - d[k][iMin2]
	for i0, dij := range dj0 {
		if wt2 := dij + djk - dk0[i]; wt2 < wtMin2 {
			wtMin2 = wt2
			iMin2 = i0
		}
	}
	for i1, dij := range dj1 {
		if wt2 := dij + djk - dk1[i1]; wt2 < wtMin2 {
			wtMin2 = wt2
			iMin2 = len(dj0) + 3 + i1
		}
	}
	return wtMin2 / 2, iMin2, k
}

// limbWeightSubMatrix finds the weight of the edge to a leaf (a limb) in
// a phylogenic tree corresponding to a submatrix of d.
//
// Argument j is an index of d representing a leaf node.  The submatrix
// considered is d[:j+1][:j+1].  This variation of limbWeight allows
// simpler and more efficient code.
//
// Return values are the same as for limbWeight.
func (d DistanceMatrix) limbWeightSubMatrix(j int) (wt float64, i, k int) {
	k = j - 1
	dj := d[j]
	dk := d[k]
	djk := dj[k]
	iMin2 := k - 1
	wtMin2 := dj[iMin2] + djk - dk[iMin2]
	for i, dij := range dj[:iMin2] {
		if wt := dij + djk - dk[i]; wt < wtMin2 {
			wtMin2 = wt
			iMin2 = i
		}
	}
	return wtMin2 / 2, iMin2, k
}

// AdditiveTree constructs an unrooted tree from an additive distance matrix.
//
// DistanceMatrix d must be additive.  Use provably additive matrices or
// use DistanceMatrix.Additive() to verify the additive property.
//
// Result is an unrooted tree, not necessarily binary, as an undirected graph.
// The first len(d) nodes are the leaves represented by the distance matrix.
// Internal nodes follow.
//
// Time complexity is O(n^2) in the number of leaves.
func (d DistanceMatrix) AdditiveTree() (u graph.LabeledUndirected, edgeWts []float64) {
	// interpretation of the presented recursive algorithm.  ideas of
	// things to try:  1: construct result as a parent list rather than
	// a child tree.  2: drop the recursion.  3. make tree always binary.
	t := make(graph.LabeledAdjacencyList, len(d), len(d)+len(d)-2)
	var ap func(int)
	ap = func(n int) {
		if n == 1 {
			edgeWts = []float64{d[0][1]}
			t[0] = []graph.Half{{1, 0}}
			t[1] = []graph.Half{{0, 0}}
			return
		}
		nLen, i, k := d.limbWeightSubMatrix(n)
		x := d[i][n] - nLen
		ap(n - 1)
		// f() finds and returns connection node v.
		// method: df search to find i from k, find connection point on the
		// way out.
		// create connection node v if needed, return v if found, -1 if not.
		vis := bits.New(len(t))
		var f func(n graph.NI) graph.NI
		f = func(n graph.NI) graph.NI {
			if int(n) == i {
				return n
			}
			vis.SetBit(int(n), 1)
			for tx, to := range t[n] {
				if vis.Bit(int(to.To)) == 1 {
					continue
				}
				p := f(to.To)
				switch {
				case p < 0: // not found yet
					continue
				case x == 0: // p is connection node
					return p
				case x < edgeWts[to.Label]: // new node at dist x from to.To
					// plan is to recycle the existing half edges between
					// n and to.To to go to new node v.  The edge(n, v)
					// gets to keep the recycled edge label with weight
					// reduced by x.  The edge(to.To, v) gets a new edge label
					// with weight x.

					v := graph.NI(len(t))  // new node
					t[n][tx].To = v        // redirect half
					edgeWts[to.Label] -= x // reduce wt

					y := graph.LI(len(edgeWts)) // new label for edge(to.To, v)
					edgeWts = append(edgeWts, x)
					// now find reciprocal half from to.To back to n
					for fx, from := range t[to.To] {
						if from.To == n { // here it is
							// recycle it to go to v now.
							t[to.To][fx] = graph.Half{v, y}
							break
						}
					}
					t = append(t, []graph.Half{{n, to.Label}, {to.To, y}})
					x = 0
					return v
				default: // continue back out
					x -= edgeWts[to.Label]
					return n
				}
			}
			return -1
		}
		vis.ClearAll()
		v := f(graph.NI(k))
		y := graph.LI(len(edgeWts))
		edgeWts = append(edgeWts, nLen)
		t[n] = []graph.Half{{v, y}}
		t[v] = append(t[v], graph.Half{graph.NI(n), y})
	}
	ap(len(d) - 1)
	return graph.LabeledUndirected{t}, edgeWts
}

// RAMatrix constructs a random additive distance matrix.
//
// Argument n is the size of the DistanceMatrix to reutrn.
func RandomAdditiveMatrix(n int) DistanceMatrix {
	pl := randomUTree(n)
	da := make([]struct { // distance annotation of parent list
		leng int     // path length
		wt   float64 // edge weight to parent
		dist float64 // distance to root
	}, len(pl))
	for i := range da {
		da[i].wt = 10 + float64(rand.Intn(90))
	}
	var f func(int) (int, float64)
	f = func(n int) (int, float64) {
		switch {
		case n == len(pl):
			return 1, 0
		case da[n].leng > 0:
			return da[n].leng, da[n].dist
		}
		leng, dist := f(pl[n])
		leng++
		dist += da[n].wt
		da[n].leng = leng
		da[n].dist = dist
		return leng, dist
	}
	for leaf := 0; leaf < n; leaf++ {
		f(leaf)
	}
	// distance between leaves in annotated parent list.
	lldist := func(l1, l2 int) (d float64) {
		// make l1 the leaf with the longer path
		if da[l1].leng < da[l2].leng {
			l1, l2 = l2, l1
		}
		// accumulate l1 distance to same tree height as l2
		len2 := da[l2].leng
		for da[l1].leng > len2 {
			d += da[l1].wt
			l1 = pl[l1]
		}
		// accumulate d1, d2 until l1, l2 are the same node
		for l1 != l2 {
			d += da[l1].wt + da[l2].wt
			l1 = pl[l1]
			l2 = pl[l2]
		}
		return
	}
	// build distance matrix
	d := make(DistanceMatrix, n)
	for i := range d {
		di := make([]float64, n)
		for j := 0; j < i; j++ {
			di[j] = lldist(i, j)
			d[j][i] = di[j]
		}
		d[i] = di
	}
	return d
}

// Ultrametric labels nodes in the tree returned by DistanceMatrix.Ultrametric.
type Ultrametric struct {
	Weight float64 // edge weight from parent (the evolutionary distance)
	Age    float64 // age (height above leaves)
}

// DAVG, DMIN constants for argument to Ultrametric.
const (
	DAVG = iota // UPGMA (average) custer distance metric
	DMIN        // single linkage (minimum) cluster distance metric
)

// Ultrametric constructs a rooted ultrametric binary tree from
// DistanceMatrix dm.
//
// The function is destructive on dm.
//
// The tree result is returned as a parent list, A list of nodes where each
// points to its parent.  Leaves of the tree are represented by elements
// 0:len(dm).  Age only increases in the list.  The root is the last element
// in the list.  Having no logical parent, the root will have parent = -1 and
// Weight = NaN.  It will also have NLeaves = len(dm).
//
// See also UltrametricD.
func (dm DistanceMatrix) Ultrametric(cdf int) (graph.FromList, []Ultrametric) {
	return dm.Clone().UltrametricD(cdf)
}

// UltrametricD is the same as Ultrametric but is destructive on the receiver.
//
// It saves a little memory if you have no further use for the distance matrix.
func (dm DistanceMatrix) UltrametricD(cdf int) (graph.FromList, []Ultrametric) {
	pl := make([]graph.PathEnd, len(dm)) // the parent-list
	ul := make([]Ultrametric, len(dm))   // labels for the parent-list
	for i := range pl {
		// "initial isolated nodes"
		pl[i] = graph.PathEnd{
			From: -1,
			Len:  1,
		}
		ul[i] = Ultrametric{Weight: math.NaN(), Age: 0}
	}

	// clusters is the list of clusters available for merging.  it starts
	// with all leaf nodes and is reduced in length as clusters are merged.
	// values represent distance matrix indexes
	clusters := make([]int, len(dm))
	for i := range dm {
		clusters[i] = i
	}
	// cx converts a distance matrix index to a node number
	cx := make([]graph.NI, len(dm))
	for i := range dm {
		cx[i] = graph.NI(i)
	}

	// extra workspace for DMIN
	var cl [][]int
	if cdf == DMIN {
		cl = make([][]int, len(dm)*2-1)
		for i := range dm {
			cl[i] = []int{i}
		}
	}

	for {
		d1, d2, cl2 := dm.closest(clusters)
		c1 := cx[d1] // cluster (node) numbers
		c2 := cx[d2]
		di1 := dm[d1] // rows in distance matrix
		di2 := dm[d2]
		m1 := pl[c1].Len // number of leaves in each cluster
		m2 := pl[c2].Len
		m3 := m1 + m2 // total number of leaves for new cluster

		// create node here, initial values come from d1, d2
		parent := graph.NI(len(pl))
		age := di2[d1] / 2
		pl = append(pl, graph.PathEnd{
			From: -1,
			Len:  m3,
		})
		ul = append(ul, Ultrametric{
			Weight: math.NaN(),
			Age:    age,
		})
		pl[c1].From = parent
		pl[c2].From = parent
		ul[c1].Weight = age - ul[c1].Age
		ul[c2].Weight = age - ul[c2].Age

		if len(clusters) == 2 {
			break
		}

		cx[d1] = parent
		// replace d1 with new computed distance
		switch cdf {
		case DAVG:
			mag1 := float64(m1)
			mag2 := float64(m2)
			invMag := 1 / float64(m3)
			for _, j := range clusters {
				dij := di1[j]
				if j != d1 {
					d := (dij*mag1 + di2[j]*mag2) * invMag
					di1[j] = d
					dm[j][d1] = d
				}
			}
		case DMIN:
			for _, j := range clusters {
				dj1 := di1[j]
				if dj2 := di2[j]; dj2 < dj1 {
					di1[j] = dj2
				} else {
					dm[j][d1] = dj1
				}
			}
		default:
			panic("Ultrametric: invalid distance function")
		}
		// d1 has been replaced, delete d2
		last := len(clusters) - 1
		clusters[cl2] = clusters[last]
		clusters = clusters[:last]
	}
	return graph.FromList{Paths: pl}, ul
}

// closest clusters (min value in d) among only those clusters (d indexes)
// listed in argument `clusters`
// return smaller index (iMin) first
// also return index into cluster list of jMin so it can be deleted later.
func (dm DistanceMatrix) closest(clusters []int) (iMin, jMin, cj int) {
	min := math.Inf(1)
	iMin = -1
	jMin = -1
	for _, i := range clusters {
		for c, j := range clusters {
			if i < j {
				if d := dm[i][j]; d < min {
					min = d
					iMin = i
					jMin = j
					cj = c
				}
			}
		}
	}
	return
}

// Cut partitions leaf nodes of an ultrametric tree into k clusters.
//
// A UList of length l represents a tree with nLeaves = (l+1)/2 leaves.
// Cut returns an k-partition of 0:nLeaves.  Each partition corresponds to
// a subtree of u.  Because of the increasing age property of a UList,
// the parents of the roots of the k subtrees will be the last k-1
// elements of the list, that is the parents will be >= len(u)-(k-1).
/*
func (u UList) Cut(k int) (clusters [][]int) {
	nLeaves := (len(u) + 1) / 2
	if k > nLeaves {
		k = nLeaves
	}
	clusters = make([][]int, 0, k) // return value has k clusters
	cut := len(u) - (k - 1)
	c := make([][]int, cut) // working data uses more though
	for l := range c[:nLeaves] {
		c[l] = []int{l}
	}
	for i, ui := range u[:cut] {
		uiP := ui.Parent
		if uiP >= cut {
			clusters = append(clusters, c[i])
		} else {
			c[uiP] = append(c[uiP], c[i]...)
		}
	}
	return
}
*/

// NeighborJoin constructs an unrooted tree from a distance matrix using the
// neighbor joining algorithm.
//
// The tree is returned as an undirected graph and a weight list.  Edges of
// the graph are labeled as indexes into the weight list.  Leaves of the
// the tree will be graph node 0:len(dm).
//
// See also NeighborJoinD.
func (dm DistanceMatrix) NeighborJoin() (u graph.LabeledUndirected, wt []float64) {
	return dm.Clone().NeighborJoinD()
}

// NeighborJoinD is the same as NeighborJoin but is destructive on the receiver.
//
// It saves a little memory if you have no further use for the distance matrix.
func (dm DistanceMatrix) NeighborJoinD() (u graph.LabeledUndirected, wt []float64) {
	td := make([]float64, len(dm))  // total-distance vector
	nx := make([]graph.NI, len(dm)) // node number corresponding to dist matrix index
	for i := range dm {
		nx[i] = graph.NI(i)
	}

	// closest clusters (min value in dm)
	// return smaller index (j) first
	closest := func() (jMin, iMin int) {
		min := math.Inf(1)
		iMin = -1
		jMin = -1
		for i := 1; i < len(dm); i++ {
			for j := 0; j < i; j++ {
				d := float64(len(dm)-2)*dm[i][j] - td[i] - td[j]
				if d < min {
					min = d
					iMin = i
					jMin = j
				}
			}
		}
		return
	}

	// wt is edge weight from parent (limb length)
	var tree graph.LabeledAdjacencyList
	var nj func(graph.NI)
	nj = func(m graph.NI) { // m is next internal node number
		if len(dm) == 2 {
			wt = make([]float64, 1, m-1)
			wt[0] = dm[0][1]
			tree = make(graph.LabeledAdjacencyList, m)
			n0 := nx[0]
			n1 := nx[1]
			tree[n0] = []graph.Half{{To: n1}}
			tree[n1] = []graph.Half{{To: n0}}
			return
		}
		// compute or recompute TotalDistance
		for k, dk := range dm {
			t := 0.
			for _, d := range dk {
				t += d
			}
			td[k] = t
		}
		d1, d2 := closest()
		Δ := (td[d2] - td[d1]) / float64(len(dm)-2)
		d21 := dm[d2][d1]
		ll2 := .5 * (d21 + Δ)
		ll1 := .5 * (d21 - Δ)
		n1 := nx[d1]
		n2 := nx[d2]

		di1 := dm[d1] // rows in distance matrix
		di2 := dm[d2]

		// replace d1 with mean distance
		for j, dij := range di1 {
			mn := .5 * (dij + di2[j] - d21)
			if j == d1 && mn != 0 {
				panic("uh uh, prolly skip this one...")
			}
			di1[j] = mn
			dm[j][d1] = mn
		}

		// d1 has been replaced, delete d2
		copy(dm[d2:], dm[d2+1:])
		dm = dm[:len(dm)-1]
		for i, di := range dm {
			copy(di[d2:], di[d2+1:])
			dm[i] = di[:len(di)-1]
		}
		nx[d1] = m
		copy(nx[d2:], nx[d2+1:])
		nx = nx[:len(dm)]

		// recurse
		nj(m + 1)

		// join limbs to tree
		wx1 := graph.LI(len(wt))
		wx2 := wx1 + 1
		wt = append(wt, ll1, ll2)
		tree[m] = append(tree[m],
			graph.Half{n1, wx1},
			graph.Half{n2, wx2})
		tree[n1] = append(tree[n1], graph.Half{m, wx1})
		tree[n2] = append(tree[n2], graph.Half{m, wx2})
		return
	}
	nj(graph.NI(len(dm)))
	return graph.LabeledUndirected{tree}, wt
}
