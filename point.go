// Author Sonia Keys 2012
// Public domain.

package cluster

import "math"

// Point is an n-dimensional point in Euclidean space.
type Point []float64

func (p Point) Clear() {
	for i := range p {
		p[i] = 0
	}
}

// Add, element-wise += on a Point.
func (p1 Point) Add(p2 Point) {
	for i, x2 := range p2 {
		p1[i] += x2
	}
}

// Mul, scalar multiply on a Point.
func (p Point) Mul(s float64) {
	for i := range p {
		p[i] *= s
	}
}

// SetMean, set p to the mean of pts.
func (p Point) SetMean(pts []Point) {
	copy(p, pts[0])
	for _, p2 := range pts {
		p.Add(p2)
	}
	p.Mul(1 / float64(len(pts)))
}

// Sqd, square of Euclidean distance between Points.
func (p1 Point) Sqd(p2 Point) (ssq float64) {
	for i, x1 := range p1 {
		d := x1 - p2[i]
		ssq += d * d
	}
	return
}

// NearestSqd finds the point nearest the receiver out of a list of points.
//
// Euclidean distance by Sqd.  Return values are the index of the nearest
// point and the square of the distance from the receiver to the nearest point.
func (p Point) NearestSqd(pts []Point) (int, float64) {
	iMin := 0
	sqdMin := p.Sqd(pts[0])
	for i, p2 := range pts[1:] {
		if sqd := p.Sqd(p2); sqd < sqdMin {
			sqdMin = sqd
			iMin = i + 1
		}
	}
	return iMin, sqdMin
}

func (x Point) Pearson(y Point) float64 {
	var μx, μy float64
	for i, xi := range x {
		μx += xi
		μy += y[i]
	}
	μx /= float64(len(x))
	μy /= float64(len(x))
	var sn, sx, sy float64
	for i, xi := range x {
		dx := xi - μx
		dy := y[i] - μy
		sn += dx * dy
		sx += dx * dx
		sy += dy * dy
	}
	return sn / math.Sqrt(sx*sy)
}
