// Author Sonia Keys 2015
// Public domain.

// Cluster contains a few clustering methods including K-Means++ and UPGMA.
//
// K-Means
//
// K-Means operates on an N-dimensional point type.  Three initializers are
// provided, "++", random, and first points in list.  A K-means++ wrapper is
// provided as a convenience.
//
// Expectation Maximization
//
// A soft K-Means variant uses expectation maximization.  This also operates
// on the same N-dimensional point type.
//
// Hierarchical
//
// The hierarchical methods here take a distance matrix as input.
// The method Ultrametric can perform either UPGMA or
// single-linkage clustering and produce a rooted ultrametric tree.
// Methods AdditiveTree and NeighborJoin produce unrooted binary trees.
//
// Clique approximation
//
// CAST stands for Cluster Affinity Search Technique.  It clusters points
// where linking points with a similarity threshold would approximate a
// clique graph.
//
// Utility functions
//
// A few methods are exported for the Point type, including a Pearson
// correlation coefficient function useful for constructing similarity
// matrices.  Also some data validation methods, a random tree generator
// and a random distance matrix generator.
package cluster
