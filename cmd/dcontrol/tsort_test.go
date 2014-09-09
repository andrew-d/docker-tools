package main

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TestIterations = 10000

func TestRandomGraphs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping random testing in short mode")
	}

	// Get a random seed.
	seed := rand.Int63()
	fmt.Printf("Seed = %d\n", seed)

	source := rand.NewSource(seed)
	rng := rand.New(source)

	for cnt := 0; cnt < TestIterations; cnt++ {
		numNodes := rng.Intn(100-10) + 10
		addedDeps := 0
		containers := make([]*Container, numNodes)

		// We want to generate a random graph with no cycles.  To do this, one
		// simple way is to generate random dependencies that can only point
		// "forward" in the graph, and thus in the list of containers.
		for i := 0; i < numNodes; i++ {
			containers[i] = &Container{
				Name:         fmt.Sprintf("container%d", i),
				Dependencies: []string{},
			}

			numDeps := rng.Intn(10)
			seenDeps := make(map[int]struct{}, numDeps)

			for j := 0; j < numDeps; j++ {
				// Select a random number in the range (i, numNodes]
				depNum := rng.Intn(numNodes-i) + i

				// Insert it as a dependency only if we haven't already.
				_, seen := seenDeps[depNum]

				if !seen && i != depNum {
					containers[i].Dependencies = append(
						containers[i].Dependencies,
						fmt.Sprintf("container%d", depNum),
					)
					seenDeps[depNum] = struct{}{}
					addedDeps++
				}
			}
		}

		//fmt.Printf("numNodes = %d, addedDeps = %d\n", numNodes, addedDeps)

		// Good.  Topologically sort this graph.
		toposort, err := TopoSortContainers(containers)
		assert.NoError(t, err, "Expected no cycles in the graph")

		// Validate the topologically-sorted graph by walking through the sort,
		// and marking each node 'done' as we reach it.  For each node, all of
		// its dependencies must be marked 'done' by the time we reach it.
		done := make(map[string]bool)
		for _, idx := range toposort {
			node := containers[idx]

			// Validate dependencies
			for i, dep := range node.Dependencies {
				if !done[dep] {
					t.Fatalf("Node %d: dependency %d (%s) not done", idx, i, dep)
				}
			}

			done[node.Name] = true
		}
	}
}
