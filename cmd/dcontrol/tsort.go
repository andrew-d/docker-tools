package main

import (
	"fmt"
	"log"
)

var _ = log.Println

// TopoSortContainers uses the Kahn algorithm for topological sorting (adapted
// from Wikipedia and RosettaCode)
// See:
//		http://rosettacode.org/wiki/Topological_sort#Go
//		https://en.wikipedia.org/wiki/Topological_sorting
func TopoSortContainers(containers []*Container) ([]int, error) {
	// Convert names to indexes.
	indexes := make(map[string]int, len(containers))
	for i, c := range containers {
		indexes[c.Name] = i
	}

	// Convert our array into a graph.  The format of the edges map is:
	//	  edges[x] is a list of all edges that go from x --> y
	//
	// For the topological sort, an edge from x --> y indicates that y depends
	// on x - i.e. x must be done first.
	edges := make(map[int][]int)

	// We also store the 'degree' - i.e. for a node n, the number of other
	// nodes that depend on this node.
	degree := make(map[int]int)

	for ci, c := range containers {
		for _, dep := range c.Dependencies {
			if dep == c.Name {
				return nil, fmt.Errorf("Container '%s' depends on itself", dep)
			}

			// Ensure the dependency exists.
			if _, ok := indexes[dep]; !ok {
				return nil, fmt.Errorf("Dependency '%s' for container '%s' does not exist",
					dep, c.Name)
			}

			edges[indexes[dep]] = append(edges[indexes[dep]], ci)
			degree[ci]++
		}
	}

	// Find a list of "start nodes" that have no incoming edges - i.e. they
	// have no dependencies.
	S := []int{}
	for i := range containers {
		if degree[i] == 0 {
			S = append(S, i)
		}
	}

	// L is the list that will contain sorted elements.
	L := []int{}

	for len(S) > 0 {
		last := len(S) - 1 // "remove a node n from S"
		n := S[last]
		S = S[:last]

		L = append(L, n) // "add n to tail of L"
		for _, m := range edges[n] {
			// WP pseudo code reads "for each node m..." but it means for each
			// node m *remaining in the graph.*  So, "remaining in the graph"
			// for us means degree[m] > 0.
			if degree[m] > 0 {
				degree[m]--         // "remove edge from the graph"
				if degree[m] == 0 { // if "m has no other incoming edges"
					S = append(S, m) // "insert m into S"
				}
			}
		}
	}

	// "If graph has edges," for us means a value in degree is > 0.
	cyclic := []int{}
	for c, in := range degree {
		if in > 0 {
			// recover cyclic nodes
			for _, nb := range edges[c] {
				if degree[nb] > 0 {
					cyclic = append(cyclic, c)
					break
				}
			}
		}
	}
	if len(cyclic) > 0 {
		msg := containers[cyclic[0]].Name
		for _, c := range cyclic[1:] {
			msg += ", " + containers[c].Name
		}

		return nil, fmt.Errorf("Cycle detected among: %s", msg)
	}
	return L, nil
}
