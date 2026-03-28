package core

import "fmt"

// TopologicalSortTables sorts tables based on foreign key dependencies using
// Kahn's algorithm (BFS). Tables with no dependencies come first so that
// parent rows are inserted before child rows, preventing FK violation errors.
//
// Only FK references to tables within the provided slice are considered.
// Self-referencing FKs are ignored for ordering purposes.
// Returns an error if a circular dependency is detected.
func TopologicalSortTables(tables []Table) ([]Table, error) {
	tableMap := make(map[string]*Table, len(tables))
	tableSet := make(map[string]struct{}, len(tables))
	for i := range tables {
		key := tables[i].Name
		tableMap[key] = &tables[i]
		tableSet[key] = struct{}{}
	}

	// Build adjacency list and in-degree map.
	// Edge: referencedTable -> dependentTable
	inDegree := make(map[string]int, len(tables))
	adjacency := make(map[string][]string, len(tables))

	for _, t := range tables {
		if _, ok := inDegree[t.Name]; !ok {
			inDegree[t.Name] = 0
		}
		for _, fk := range t.ForeignKeys {
			ref := fk.ReferencedTable

			// Skip self-references and references to tables outside this set.
			if ref == t.Name {
				continue
			}
			if _, exists := tableSet[ref]; !exists {
				continue
			}

			adjacency[ref] = append(adjacency[ref], t.Name)
			inDegree[t.Name]++
		}
	}

	// Kahn's algorithm: start with all zero in-degree nodes.
	queue := make([]string, 0)
	for _, t := range tables {
		if inDegree[t.Name] == 0 {
			queue = append(queue, t.Name)
		}
	}

	sorted := make([]Table, 0, len(tables))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, *tableMap[current])

		for _, neighbor := range adjacency[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(sorted) != len(tables) {
		// Collect tables involved in the cycle.
		cycled := make([]string, 0)
		for name, deg := range inDegree {
			if deg > 0 {
				cycled = append(cycled, name)
			}
		}
		return nil, fmt.Errorf("circular foreign key dependency detected among tables: %v", cycled)
	}

	return sorted, nil
}
