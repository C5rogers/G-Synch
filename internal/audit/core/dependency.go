package core

import (
	"fmt"
	"sort"
)

// DependencyGroup represents a strongly connected set of tables in dependency order.
// Cyclic groups need special handling during sync because no single insert order
// can satisfy all foreign keys without deferring constraints or splitting writes.
type DependencyGroup struct {
	Tables []Table
	Cyclic bool
}

// TopologicalSortTablesPartial sorts as many tables as possible using Kahn's
// algorithm and returns the remaining tables that participate in cycles.
//
// Only FK references to tables within the provided slice are considered.
// Self-referencing FKs are ignored for ordering purposes.
func TopologicalSortTablesPartial(tables []Table) ([]Table, []Table) {
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

	if len(sorted) == len(tables) {
		return sorted, nil
	}

	cycled := make([]string, 0)
	for name, deg := range inDegree {
		if deg > 0 {
			cycled = append(cycled, name)
		}
	}
	sort.Strings(cycled)

	cyclicTables := make([]Table, 0, len(cycled))
	for _, name := range cycled {
		if table, ok := tableMap[name]; ok {
			cyclicTables = append(cyclicTables, *table)
		}
	}

	return sorted, cyclicTables
}

// TopologicalSortTables keeps the original strict behavior for callers that
// need to fail fast when a cycle exists.
func TopologicalSortTables(tables []Table) ([]Table, error) {
	sorted, cyclicTables := TopologicalSortTablesPartial(tables)
	if len(cyclicTables) > 0 {
		cycled := make([]string, 0, len(cyclicTables))
		for _, table := range cyclicTables {
			cycled = append(cycled, table.Name)
		}
		return nil, fmt.Errorf("circular foreign key dependency detected among tables: %v", cycled)
	}

	return sorted, nil
}

// BuildDependencyPlan returns dependency groups in topological order.
// Acyclic groups contain a single table. Cyclic groups contain one or more
// tables that must be handled together.
func BuildDependencyPlan(tables []Table) ([]DependencyGroup, error) {
	if len(tables) == 0 {
		return nil, nil
	}

	tableMap := make(map[string]*Table, len(tables))
	tableSet := make(map[string]struct{}, len(tables))
	for i := range tables {
		key := tables[i].Name
		tableMap[key] = &tables[i]
		tableSet[key] = struct{}{}
	}

	adjacency := make(map[string][]string, len(tables))
	for _, t := range tables {
		for _, fk := range t.ForeignKeys {
			ref := fk.ReferencedTable
			if _, exists := tableSet[ref]; !exists {
				continue
			}
			adjacency[ref] = append(adjacency[ref], t.Name)
		}
	}

	componentNames := stronglyConnectedComponents(tables, adjacency)
	componentIndex := make(map[string]int, len(tables))
	for i, component := range componentNames {
		for _, name := range component {
			componentIndex[name] = i
		}
	}

	componentGraph := make(map[int]map[int]struct{}, len(componentNames))
	inDegree := make(map[int]int, len(componentNames))
	for i := range componentNames {
		inDegree[i] = 0
	}

	for _, t := range tables {
		from := componentIndex[t.Name]
		for _, fk := range t.ForeignKeys {
			ref := fk.ReferencedTable
			if _, exists := tableSet[ref]; !exists {
				continue
			}
			to := componentIndex[ref]
			if from == to {
				continue
			}
			if componentGraph[to] == nil {
				componentGraph[to] = map[int]struct{}{}
			}
			if _, seen := componentGraph[to][from]; seen {
				continue
			}
			componentGraph[to][from] = struct{}{}
			inDegree[from]++
		}
	}

	queue := make([]int, 0)
	for i, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, i)
		}
	}
	sort.Ints(queue)

	orderedComponents := make([]int, 0, len(componentNames))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		orderedComponents = append(orderedComponents, current)

		neighbors := make([]int, 0, len(componentGraph[current]))
		for neighbor := range componentGraph[current] {
			neighbors = append(neighbors, neighbor)
		}
		sort.Ints(neighbors)
		for _, neighbor := range neighbors {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
				sort.Ints(queue)
			}
		}
	}

	if len(orderedComponents) != len(componentNames) {
		return nil, fmt.Errorf("unable to build dependency plan for all tables")
	}

	plan := make([]DependencyGroup, 0, len(orderedComponents))
	for _, componentIdx := range orderedComponents {
		names := append([]string(nil), componentNames[componentIdx]...)
		sort.Strings(names)

		groupTables := make([]Table, 0, len(names))
		cyclic := len(names) > 1
		for _, name := range names {
			if table, ok := tableMap[name]; ok {
				groupTables = append(groupTables, *table)
				if !cyclic {
					for _, fk := range table.ForeignKeys {
						if fk.ReferencedTable == table.Name {
							cyclic = true
							break
						}
					}
				}
			}
		}

		plan = append(plan, DependencyGroup{
			Tables: groupTables,
			Cyclic: cyclic,
		})
	}

	return plan, nil
}

func stronglyConnectedComponents(tables []Table, adjacency map[string][]string) [][]string {
	index := 0
	stack := make([]string, 0, len(tables))
	onStack := make(map[string]bool, len(tables))
	indices := make(map[string]int, len(tables))
	lowlink := make(map[string]int, len(tables))
	components := make([][]string, 0, len(tables))

	var visit func(string)
	visit = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		neighbors := append([]string(nil), adjacency[v]...)
		sort.Strings(neighbors)
		for _, w := range neighbors {
			if _, seen := indices[w]; !seen {
				visit(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] && indices[w] < lowlink[v] {
				lowlink[v] = indices[w]
			}
		}

		if lowlink[v] == indices[v] {
			component := make([]string, 0)
			for {
				n := len(stack) - 1
				w := stack[n]
				stack = stack[:n]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			components = append(components, component)
		}
	}

	names := make([]string, 0, len(tables))
	for _, table := range tables {
		names = append(names, table.Name)
	}
	sort.Strings(names)

	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		if _, ok := indices[name]; !ok {
			visit(name)
		}
	}

	return components
}
