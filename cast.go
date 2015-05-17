// Public domain.

package cluster

// SimilarityMatrix is a square matrix reflecting similarity of samples
// represented by the two indexes.
type SimilarityMatrix [][]float64

// CAST implements "Cluster Affinity Search Technique"
//
// Argument v is a similarity threshold.  If a graph is constructed with
// nodes of the similarity matrix indices and edges where similarity is
// at least threshold v, then the algorithm clusters nodes that approximate
// graph cliques.  The number of clusters output is not fixed.  Low thresholds
// produce few clusters.  Higher thresholds partition the set into more
// clusters.
//
// The return value is a partition of the indices 0:len(sim).
func (sim SimilarityMatrix) CAST(v float64) (clusters [][]int) {
	g := make([][]bool, len(sim)) // graph
	for i, si := range sim {
		gi := make([]bool, len(si))
		for j, sij := range si {
			gi[j] = i != j && sij > v
		}
		g[i] = gi
	}
	in := map[int]bool{} // set of nodes in graph
	for i := range g {   // (initially all nodes are in graph)
		in[i] = true
	}
	for len(in) > 0 {
		// pick node of max degree
		maxDeg := -1
		maxI := -1
		for i := range in {
			deg := 0
			for j := range in {
				if g[i][j] {
					deg++
				}
			}
			if deg > maxDeg {
				maxDeg = deg
				maxI = i
			}
		}
		c := map[int]bool{maxI: true} // cluster with single node
		for {
			vCloseI := -1    // nearest v-close i not in c
			vDistI := -1     // farthest v-distant i in c
			vCloseSim := -2. // not close
			vDistSim := 2.   // close
			for i := range in {
				s := sim.cSim(i, c)
				if s > v {
					// it's v-close
					if !c[i] && s > vCloseSim { // and not in c
						vCloseSim = s
						vCloseI = i
					}
				} else {
					// it's v-distant
					if c[i] && s < vDistSim { // and in c
						vDistSim = s
						vDistI = i
					}
				}
			}
			if vCloseI < 0 && vDistI < 0 {
				break
			}
			if vCloseI >= 0 {
				c[vCloseI] = true
			}
			if vDistI >= 0 {
				delete(c, vDistI)
			}
		}
		cs := make([]int, len(c))
		i := 0
		for n := range c {
			cs[i] = n
			i++
			delete(in, n)
		}
		clusters = append(clusters, cs)
	}
	return
}

func (sim SimilarityMatrix) cSim(i int, c map[int]bool) float64 {
	s := 0.
	for j := range c {
		s += sim[i][j]
	}
	return s / float64(len(c))
}

// NewPearsonSim constructs an n×n similarity matrix where n is len(exp).
func NewPearsonSim(exp []Point) SimilarityMatrix {
	sim := make(SimilarityMatrix, len(exp))
	for i := range sim {
		si := make([]float64, len(exp))
		for j := 0; j < i; j++ {
			c := exp[i].Pearson(exp[j])
			si[j] = c
			sim[j][i] = c
		}
		si[i] = 1
		sim[i] = si
	}
	return sim
}
