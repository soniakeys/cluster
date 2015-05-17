// Public domain.

package cluster_test

import (
	"fmt"
	"sort"

	"github.com/soniakeys/cluster"
)

func ExampleSimilarityMatrix_CAST() {
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
	m := cluster.NewEuclideanDist(exp)
	// make similarity matrix by subtracting distance matrix from a constant
	for _, row := range m {
		for c, d := range row {
			row[c] = 20 - d
		}
	}
	sim := cluster.SimilarityMatrix(m)
	// threshold of 15 reproduces clusters of UList.Cut example.
	for _, c := range sim.CAST(15) {
		sort.Ints(c)
		for _, x := range c {
			fmt.Printf("%d: %g\n", x, exp[x])
		}
		fmt.Println()
	}
	// Output:
	// 1: [10 0 9]
	// 3: [9.5 0.5 8.5]
	// 8: [9.7 2 9]
	// 9: [10.2 1 9.2]
	//
	// 2: [4 8.5 3]
	// 4: [4.5 8.5 2.5]
	// 7: [3.7 8.7 2]
	//
	// 0: [10 8 10]
	// 5: [10.5 9 12]
	//
	// 6: [5 8.5 11]
}
