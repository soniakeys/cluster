package kmpp

import (
	"math"
	"math/rand"
)

type R2 struct {
	X, Y float64
}

type R2c struct {
	R2
	C int // cluster number
}

// Kmpp, K-means++.
func Kmpp(k int, data []R2c) {
	KMeans(data, kmppSeeds(k, data))
}

// kmppSeeds is the ++ part.
// It generates the initial means for the k-means algorithm.
func kmppSeeds(k int, data []R2c) []R2 {
	s := make([]R2, k)
	s[0] = data[rand.Intn(len(data))].R2
	d2 := make([]float64, len(data))
	for i := 1; i < k; i++ {
		var sum float64
		for j, p := range data {
			_, dMin := nearest(p, s[:i])
			d2[j] = dMin * dMin
			sum += d2[j]
		}
		target := rand.Float64() * sum
		j := 0
		for sum = d2[0]; sum < target; sum += d2[j] {
			j++
		}
		s[i] = data[j].R2
	}
	return s
}

// nearest finds the nearest mean to a given point.
// return values are the index of the nearest mean, and the distance from
// the point to the mean.
func nearest(p R2c, mean []R2) (int, float64) {
	iMin := 0
	dMin := math.Hypot(p.X-mean[0].X, p.Y-mean[0].Y)
	for i := 1; i < len(mean); i++ {
		d := math.Hypot(p.X-mean[i].X, p.Y-mean[i].Y)
		if d < dMin {
			dMin = d
			iMin = i
		}
	}
	return iMin, dMin
}

// KMeans, Lloyd's algorithm.
func KMeans(data []R2c, mean []R2) {
	// initial assignment
	for i, p := range data {
		cMin, _ := nearest(p, mean)
		data[i].C = cMin
	}
	mLen := make([]int, len(mean))
	for {
		// update means
		for i := range mean {
			mean[i] = R2{}
			mLen[i] = 0
		}
		for _, p := range data {
			mean[p.C].X += p.X
			mean[p.C].Y += p.Y
			mLen[p.C]++
		}
		for i := range mean {
			inv := 1 / float64(mLen[i])
			mean[i].X *= inv
			mean[i].Y *= inv
		}
		// make new assignments, count changes
		var changes int
		for i, p := range data {
			if cMin, _ := nearest(p, mean); cMin != p.C {
				changes++
				data[i].C = cMin
			}
		}
		if changes == 0 {
			return
		}
	}
}
