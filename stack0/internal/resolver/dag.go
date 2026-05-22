package resolver

import (
	"fmt"

	"github.com/elfeddi/dxp/pkg/types"
)

// DAGResolver résout l'ordre d'installation via un graphe acyclique dirigé
type DAGResolver struct{}

func NewDAGResolver() *DAGResolver {
	return &DAGResolver{}
}

// Resolve retourne les connecteurs dans l'ordre d'installation
// en respectant les dépendances (DependsOn)
func (d *DAGResolver) Resolve(config *types.DxPConfig) ([]types.Connector, error) {
	// Collecter tous les connecteurs
	all := map[string]types.Connector{}
	for _, stack := range config.Stacks {
		if !stack.Enabled {
			continue
		}
		for _, c := range stack.Connectors {
			all[c.Name] = c
		}
	}

	// Tri topologique (Kahn's algorithm)
	inDegree := map[string]int{}
	for name := range all {
		inDegree[name] = 0
	}
	for _, c := range all {
		for _, dep := range c.DependsOn {
			if _, ok := all[dep]; !ok {
				return nil, fmt.Errorf("dépendance inconnue: %s → %s", c.Name, dep)
			}
			inDegree[c.Name]++
		}
	}

	// File des connecteurs sans dépendances
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	ordered := []types.Connector{}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		ordered = append(ordered, all[curr])

		// Réduire le degré des connecteurs qui dépendent de curr
		for name, c := range all {
			for _, dep := range c.DependsOn {
				if dep == curr {
					inDegree[name]--
					if inDegree[name] == 0 {
						queue = append(queue, name)
					}
				}
			}
		}
	}

	// Cycle détecté si tous les connecteurs ne sont pas ordonnés
	if len(ordered) != len(all) {
		return nil, fmt.Errorf("cycle de dépendances détecté dans dxp.yaml")
	}

	return ordered, nil
}
