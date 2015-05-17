// Public domain.

package cluster

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand"

	"github.com/soniakeys/graph"
)

// DistanceMatrix holds a distance matrix for cluster analyses.
//
// See DistanceMatrix.Valid for typical restrictions.
type DistanceMatrix [][]float64

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
	for i, di := range d {
		if len(di) != len(d) {
			return errors.New("not square")
		}
		for j, dij := range di {
			if dij < 0 {
				return fmt.Errorf("negative element: %g", dij)
			}
			if !(dij == d[j][i]) { // reversed test catches NaNs too.
				return errors.New("not symmetric")
			}
		}
		if !(di[i] == 0) {
			return errors.New("non-zero diagonal")
		}
	}
	for i, di := range d {
		for k, dk := range d[:i] {
			dik := di[k]
			for j, dij := range di {
				if dij+dk[j] < dik {
					return fmt.Errorf("triangle inequality not satisfied: "+
						"d[%d][%d] + d[%d][%d] < d[%d][%d]", i, j, j, k, i, k)
				}
			}
		}
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
func (d DistanceMatrix) AdditiveTree() (t graph.LabeledAdjacencyList, edgeWts []float64) {
	// interpretation of the presented recursive algorithm.
	// good things to try:  1: construct result as a parent list rather than
	// a child tree.  2: drop the recursion.  3. make tree always binary.
	t = make(graph.LabeledAdjacencyList, len(d)) // allocate leaves
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
		var vis big.Int
		var f func(n int) int
		f = func(n int) int {
			if n == i {
				return i
			}
			vis.SetBit(&vis, n, 1)
			for tx, to := range t[n] {
				if vis.Bit(to.To) == 1 {
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

					v := len(t)            // new node
					t[n][tx].To = v        // redirect half
					edgeWts[to.Label] -= x // reduce wt

					y := len(edgeWts) // new label for edge(to.To, v)
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
		vis = big.Int{}
		v := f(k)
		y := len(edgeWts)
		edgeWts = append(edgeWts, nLen)
		t[n] = []graph.Half{{v, y}}
		t[v] = append(t[v], graph.Half{n, y})
	}
	ap(len(d) - 1)
	return
}

// RAMatrix constructs a random additive distance matrix.
//
// Argument n is the size of the DistanceMatrix to reutrn.
func RandomAdditiveMatrix(n int) DistanceMatrix {
	pl := graph.RandomUTree(n)
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

// []Ultrametric is a return type from DistanceMatrix.Ultrametric.
type Ultrametric struct {
	Parent  int     // parent node number, index of parent in parent list
	Weight  float64 // edge weight from parent (the evolutionary distance)
	Age     float64 // age (height above leaves)
	NLeaves int     // number of leaves at or below this node
}

type UList []Ultrametric

// DAVG, DMIN constants for argument to Ultrametric.
const (
	DAVG = iota // UPGMA (average) custer distance metric
	DMIN        // single linkage (minimum) cluster distance metric
)

// Ultrametric constructs a rooted ultrametric binary tree from
// DistanceMatrix dm.
//
// The tree result is returned as a parent list, A list of nodes where each
// points to its parent.  Leaves of the tree are represented by elements
// 0:len(dm).  Age only increases in the list.  The root is the last element
// in the list.  Having no logical parent, the root will have parent = -1 and
// Weight = NaN.  It will also have NLeaves = len(dm).
func (dm DistanceMatrix) Ultrametric(cdf int) UList {
	pl := make([]Ultrametric, len(dm)) // the parent-list
	for i := range pl {
		pl[i] = Ultrametric{-1, math.NaN(), 0, 1} // initial isolated nodes
	}

	// cx converts a distance matrix index to a cluster index (a node number)
	cx := make([]int, len(dm))
	for i := range dm {
		cx[i] = i
	}

	// closest clusters (min value in d)
	// return smaller index (j) first
	closest := func() (min float64, jMin, iMin int) {
		min = math.Inf(1)
		iMin = -1
		jMin = -1
		for i := 1; i < len(dm); i++ {
			for j := 0; j < i; j++ {
				if d := dm[i][j]; d < min {
					min = d
					iMin = i
					jMin = j
				}
			}
		}
		return
	}
	var cl [][]int
	if cdf == DMIN {
		cl = make([][]int, len(dm)*2-1)
		for i := range dm {
			cl[i] = []int{i}
		}
	}

	for {
		_, d1, d2 := closest()

		di1 := dm[d1] // rows in distance mastrix
		di2 := dm[d2]
		c1 := cx[d1] // cluster (node) numbers
		c2 := cx[d2]
		m1 := pl[c1].NLeaves // number of leaves in each cluster
		m2 := pl[c2].NLeaves
		m3 := m1 + m2 // total number of leaves for new cluster

		// create node here, initial values come from d1, d2
		root := len(pl)
		age := di2[d1] / 2
		pl = append(pl, Ultrametric{
			Parent:  -1,
			Weight:  math.NaN(),
			Age:     age,
			NLeaves: m3,
		})
		pl[c1].Parent = root
		pl[c2].Parent = root
		pl[c1].Weight = age - pl[c1].Age
		pl[c2].Weight = age - pl[c2].Age
		cx[d1] = root

		if len(dm) == 2 {
			break
		}

		// replace d1 with mean distance
		switch cdf {
		case DAVG:
			mag1 := float64(m1)
			mag2 := float64(m2)
			invMag := 1 / float64(m3)
			for j, dij := range di1 {
				if j != d1 {
					d := (dij*mag1 + di2[j]*mag2) * invMag
					di1[j] = d
					dm[j][d1] = d
				}
			}
		case DMIN:
			for j, dj1 := range di1 {
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
		copy(dm[d2:], dm[d2+1:])
		dm = dm[:len(dm)-1]
		for i, di := range dm {
			copy(di[d2:], di[d2+1:])
			dm[i] = di[:len(di)-1]
		}
		// delete d2 from cluster index
		copy(cx[d2:], cx[d2+1:])
		cx = cx[:len(dm)]
	}
	return pl
}

// Cut partitions leaf nodes of an ultrametric tree into k clusters.
//
// A UList of length l represents a tree with nLeaves = (l+1)/2 leaves.
// Cut returns an k-partition of 0:nLeaves.  Each partition corresponds to
// a subtree of u.  Because of the increasing age property of a UList,
// the parents of the roots of the k subtrees will be the last k-1
// elements of the list, that is the parents will be >= len(u)-(k-1).
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

// NeighborJoin constructs an unrooted tree from a distance matrix using the
// neighbor joining algorithm.
//
// The tree is returned as an undirected graph and a weight list.  Edges of
// the graph are labeled as indexes into the weight list.  Leaves of the
// the tree will be graph node 0:len(dm).
func (dm DistanceMatrix) NeighborJoin() (tree graph.LabeledAdjacencyList, wt []float64) {
	// first copy dm so original is not destroyed
	dc := make(DistanceMatrix, len(dm))
	for i, di := range dm {
		dc[i] = append([]float64{}, di...)
	}
	dm = dc
	td := make([]float64, len(dm)) // total-distance vector
	nx := make([]int, len(dm))     // node number corresponding to dist matrix index
	for i := range dm {
		nx[i] = i
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
	var nj func(int)
	nj = func(m int) { // m is next internal node number
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
		wx1 := len(wt)
		wx2 := len(wt) + 1
		wt = append(wt, ll1, ll2)
		tree[m] = append(tree[m],
			graph.Half{n1, wx1},
			graph.Half{n2, wx2})
		tree[n1] = append(tree[n1], graph.Half{m, wx1})
		tree[n2] = append(tree[n2], graph.Half{m, wx2})
		return
	}
	nj(len(dm))
	return
}
