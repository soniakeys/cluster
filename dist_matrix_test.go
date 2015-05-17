// Public domain.

package cluster_test

import (
	"fmt"
	"testing"

	"github.com/soniakeys/cluster"
)

func ExampleDistanceMatrix_Validate() {
	d := cluster.DistanceMatrix{
		{0, 13, 21, 22},
		{13, 0, 12, 13},
		{21, 12, 0, 13},
		{22, 13, 13, 0},
	}
	fmt.Println(d.Validate())
	// Output:
	// <nil>
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
	for n, to := range t {
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
	pl := d.Ultrametric(cluster.DAVG)
	fmt.Println("node  parent  weight     age  leaves")
	for i, u := range pl {
		fmt.Printf(">%3d     %3d  %6.3f  %6.3f     %3d\n",
			i, u.Parent, u.Weight, u.Age, u.NLeaves)
	}
	// Output:
	// node  parent  weight     age  leaves
	// >  0       5   7.000   0.000       1
	// >  1       6   8.833   0.000       1
	// >  2       4   5.000   0.000       1
	// >  3       4   5.000   0.000       1
	// >  4       5   2.000   5.000       2
	// >  5       6   1.833   7.000       3
	// >  6      -1     NaN   8.833       4
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
	for n, to := range tree {
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
