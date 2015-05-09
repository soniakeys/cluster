// Author Sonia Keys 2012
// Public domain.

// K-means and K-means++ clustering for n-dimensional data
package kmpp

import "math/rand"

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
		cNums[i], _ = p.Nearest(centers)
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
			if cx, _ := p.Nearest(centers); cx != cNums[i] {
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
func Kmpp(points []Point, k int) (centers []Point, cNums, cCounts []int) {
	centers = kmppSeeds(points, k)
	cNums, cCounts = KMeans(points, centers)
	return
}

// kmppSeeds is the ++ part.
// It generates the initial means for the k-means algorithm.
func kmppSeeds(points []Point, k int) []Point {
	s := make([]Point, k)
	s[0] = append(Point{}, points[rand.Intn(len(points))]...)
	d2 := make([]float64, len(points))
	for i := 1; i < k; i++ {
		var sum float64
		for j, p := range points {
			_, dMin := p.Nearest(s[:i])
			d2[j] = dMin * dMin
			sum += d2[j]
		}
		target := rand.Float64() * sum
		j := 0
		for sum = d2[0]; sum < target; sum += d2[j] {
			j++
		}
		s[i] = append(Point{}, points[j]...)
	}
	return s
}
