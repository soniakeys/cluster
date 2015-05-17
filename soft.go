// Public domain.

package cluster

import "math"

// SoftKM is a soft K-means.
//
// Argument β is a "stiffness parameter" that controls how strongly points
// associate with clusters.  As β approaches ∞, SoftKM behavior approaches
// that of K-means.  β is related to a "length scale" σ where units of σ
// are those of the points argument:
//
//     β ∝ 1 / σ²
//
// Argument n is the number of iterations to perform.
//
// Return value resp[c][p] is responsibility of point p for cluster c.
func SoftKM(points, centers []Point, β float64, n int) (resp [][]float64) {
	// TODO add convergence criteria.
	m := len(points[0])
	nβ := -β
	resp = make([][]float64, len(centers))
	for i := range resp {
		resp[i] = make([]float64, len(points))
	}
	EStep := func() {
		for j, p := range points {
			sum := 0.
			for i, c := range centers {
				e := math.Exp(nβ * math.Sqrt(c.Sqd(p)))
				resp[i][j] = e
				sum += e
			}
			for i := range resp {
				resp[i][j] /= sum
			}
		}
	}
	MStep := func() {
		for i, hi := range resp {
			ci := centers[i] // put results here
			// first compute denominator
			sum := 0.
			for _, hij := range hi {
				sum += hij
			}
			f := 1 / sum
			// now compute hi dot points, (broadcast across dimensions)
			for d := 0; d < m; d++ {
				sum = 0
				for j, p := range points {
					sum += p[d] * hi[j]
				}
				ci[d] = sum * f
			}
		}
	}
	for i := 0; i < n; i++ {
		EStep()
		MStep()
	}
	return resp
}
