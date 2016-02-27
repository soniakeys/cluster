// Public domain.

package cluster_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/soniakeys/cluster"
)

func ExampleDistanceMatrix_String() {
	fmt.Println(cluster.DistanceMatrix{
		{0, 3},
		{3, 0},
	})
	fmt.Println()
	fmt.Println(cluster.DistanceMatrix{
		{.5, 1. / 3},
		{math.Inf(1), math.NaN()},
	})
	// Output:
	// [0 3]
	// [3 0]
	//
	// [0.5 0.3333333333333333]
	// [+Inf NaN]
}

func ExampleDistanceMatrix_Square() {
	d1 := cluster.DistanceMatrix{
		{0, 3},
		{3, 0},
	}
	d2 := cluster.DistanceMatrix{}
	d3 := cluster.DistanceMatrix{
		{0},
		{3, 0},
	}
	fmt.Println(d1.Square())
	fmt.Println(d2.Square())
	fmt.Println(d3.Square())
	// Output:
	// true
	// true
	// false
}

func ExampleDistanceMatrix_NonNegative() {
	d1 := cluster.DistanceMatrix{
		{0, -1}, // false
		{-1, 0},
	}
	d2 := cluster.DistanceMatrix{
		{0}, // true
		{1, 2},
	}
	d3 := cluster.DistanceMatrix{} // true (no negatives present)
	d4 := cluster.DistanceMatrix{
		{0, math.NaN()}, // true
		{3, 0},
	}
	fmt.Println(d1.NonNegative())
	fmt.Println(d2.NonNegative())
	fmt.Println(d3.NonNegative())
	fmt.Println(d4.NonNegative())
	// Outpupt:
	// false
	// true
	// true
	// true
}

func ExampleDistanceMatrix_ZeroDiagonal() {
	d1 := cluster.DistanceMatrix{
		{0, 3}, // true
		{3, 0},
	}
	d2 := cluster.DistanceMatrix{
		{0}, // true
		{7, 0},
	}
	d3 := cluster.DistanceMatrix{} // true
	d4 := cluster.DistanceMatrix{
		{0, 3},
		{3, 1e-300}, // false
	}
	fmt.Println(d1.ZeroDiagonal())
	fmt.Println(d2.ZeroDiagonal())
	fmt.Println(d3.ZeroDiagonal())
	fmt.Println(d4.ZeroDiagonal())
	// Output:
	// true
	// true
	// true
	// false
}

func ExampleDistanceMatrix_Symmetric() {
	d1 := cluster.DistanceMatrix{
		{0, 3}, // true
		{3, 0},
	}
	d2 := cluster.DistanceMatrix{
		{0, 3}, // false
		{7, 0},
	}
	d3 := cluster.DistanceMatrix{
		{0, math.NaN()}, // false (NaNs do not compare equal)
		{math.NaN(), 0},
	}
	d4 := cluster.DistanceMatrix{
		{0, 3}, // true (diagonal is not checked)
		{3, math.NaN()},
	}
	fmt.Println(d1.Symmetric())
	fmt.Println(d2.Symmetric())
	fmt.Println(d3.Symmetric())
	fmt.Println(d4.Symmetric())
	// Output:
	// true
	// false
	// false
	// true
}

func ExampleDistanceMatrix_TriangleInequality() {
	d1 := cluster.DistanceMatrix{
		{0, 13, 21, 22}, // true
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	d2 := cluster.DistanceMatrix{
		{0, 13, 21, math.NaN()}, // true
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	d3 := cluster.DistanceMatrix{}
	d4 := cluster.DistanceMatrix{
		{0, 4, 6, 1}, // false
		{4, 0, 3, 2},
		{6, 3, 0, 5},
		{1, 2, 5, 0},
	}
	fmt.Println(d1.TriangleInequality())
	fmt.Println(d2.TriangleInequality())
	fmt.Println(d3.TriangleInequality())
	fmt.Println(d4.TriangleInequality())
	// Output:
	// true 0 0 0
	// true 0 0 0
	// true 0 0 0
	// false 1 3 0
}

func ExampleDistanceMatrix_Validate() {
	d1 := cluster.DistanceMatrix{
		{0, 13, 21, 22},
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	d2 := cluster.DistanceMatrix{
		{0, 4, 6, 1}, // false
		{4, 0, 3, 2},
		{6, 3, 0, 5},
		{1, 2, 5, 0},
	}
	fmt.Println(d1.Validate())
	fmt.Println(d2.Validate())
	// Output:
	// <nil>
	// triangle inequality not satisfied: d[1][3] + d[3][0] < d[1][0]
}

func ExampleDistanceMatrix_Additive() {
	a := cluster.DistanceMatrix{
		{0, 13, 21, 22},
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	na := cluster.DistanceMatrix{
		{0, 3, 4, 3},
		{3, 0, 4, 5},
		{4, 4, 0, 2},
		{3, 5, 2, 0},
	}
	fmt.Println(a.Additive())
	fmt.Println(na.Additive())
	// Output:
	// true 0 0 0 0
	// false 3 1 0 2
}

func ExampleDistanceMatrix_AdditiveTree() {
	d := cluster.DistanceMatrix{
		{0, 13, 21, 22},
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	t, wts := d.AdditiveTree()
	for n, to := range t.LabeledAdjacencyList {
		for _, to := range to {
			fmt.Printf("%d: to %d label %d weight %g\n",
				n, to.To, to.Label, wts[to.Label])
		}
	}
	// Output:
	// 0: to 4 label 1 weight 11
	// 1: to 4 label 0 weight 2
	// 2: to 5 label 2 weight 6
	// 3: to 5 label 4 weight 7
	// 4: to 1 label 0 weight 2
	// 4: to 0 label 1 weight 11
	// 4: to 5 label 3 weight 4
	// 5: to 2 label 2 weight 6
	// 5: to 4 label 3 weight 4
	// 5: to 3 label 4 weight 7
}
func TestRandomAdditiveMatrix(t *testing.T) {
	for _, d := range []cluster.DistanceMatrix{
		cluster.RandomAdditiveMatrix(10),
		cluster.RandomAdditiveMatrix(20),
		cluster.RandomAdditiveMatrix(40),
	} {
		if ok, _, _, _, _ := d.Additive(); !ok {
			t.Fatal(len(d))
		}
	}
}

func ExampleDistanceMatrix_Ultrametric() {
	d := cluster.DistanceMatrix{
		{0, 20, 17, 11},
		{20, 0, 20, 13},
		{17, 20, 0, 10},
		{11, 13, 10, 0},
	}
	pl, ul := d.Ultrametric(cluster.DAVG)
	fmt.Println("node  leaves  parent  weight     age")
	for n, p := range pl.Paths {
		fmt.Printf(">%3d     %3d     %3d  %6.3f  %6.3f\n",
			n, p.Len, p.From, ul[n].Weight, ul[n].Age)
	}
	// Output:
	// node  leaves  parent  weight     age
	// >  0       1       5   7.000   0.000
	// >  1       1       6   8.833   0.000
	// >  2       1       4   5.000   0.000
	// >  3       1       4   5.000   0.000
	// >  4       2       5   2.000   5.000
	// >  5       3       6   1.833   7.000
	// >  6       4      -1     NaN   8.833
}

func ExampleDistanceMatrix_NeighborJoin() {
	d := cluster.DistanceMatrix{
		{0, 23, 27, 20},
		{23, 0, 30, 28},
		{27, 30, 0, 30},
		{20, 28, 30, 0},
	}
	tree, wt := d.NeighborJoin()
	fmt.Println("n1  n2  weight")
	for n, to := range tree.LabeledAdjacencyList {
		for _, h := range to {
			fmt.Printf("%d  %2d   %6.3f\n", n, h.To, wt[h.Label])
		}
	}
	// Output:
	// n1  n2  weight
	// 0   5    8.000
	// 1   4   13.500
	// 2   4   16.500
	// 3   5   12.000
	// 4   5    2.000
	// 4   1   13.500
	// 4   2   16.500
	// 5   3   12.000
	// 5   0    8.000
	// 5   4    2.000
}

/*
func ExampleUList_Cut() {
	exp := []cluster.Point{
		{10, 8, 10},
		{10, 0, 9},
		{4, 8.5, 3},
		{9.5, .5, 8.5},
		{4.5, 8.5, 2.5},
		{10.5, 9, 12},
		{5, 8.5, 11},
		{3.7, 8.7, 2},
		{9.7, 2, 9},
		{10.2, 1, 9.2},
	}
	dm := cluster.NewEuclideanDist(exp)
	u := dm.Ultrametric(cluster.DAVG)
	for _, c := range u.Cut(4) {
		for _, x := range c {
			fmt.Printf("%d: %g\n", x, exp[x])
		}
		fmt.Println()
	}
	// Output:
	// 6: [5 8.5 11]
	//
	// 7: [3.7 8.7 2]
	// 2: [4 8.5 3]
	// 4: [4.5 8.5 2.5]
	//
	// 8: [9.7 2 9]
	// 9: [10.2 1 9.2]
	// 1: [10 0 9]
	// 3: [9.5 0.5 8.5]
	//
	// 0: [10 8 10]
	// 5: [10.5 9 12]
}
*/
