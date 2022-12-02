package schedulerd

import (
	"sort"

	corev2 "github.com/sensu/core/v2"
)

// namespacedChecks is a namespaced collection of checkSets
type namespacedChecks map[string]checkSet

// Update a set of checks by namespace
// returns a slice for CheckConfigs that have been added, changed and removed
func (s namespacedChecks) Update(next []*corev2.CheckConfig) ([]*corev2.CheckConfig, []*corev2.CheckConfig, []*corev2.CheckConfig) {
	var added, changed, removed []*corev2.CheckConfig
	byNamespace := make(map[string][]*corev2.CheckConfig)
	for _, c := range next {
		byNamespace[c.Namespace] = append(byNamespace[c.Namespace], c)
	}
	for ns := range s {
		if _, ok := byNamespace[ns]; !ok {
			// removed namespace
			for _, check := range s[ns].set {
				removed = append(removed, check)
			}
			delete(s, ns)
		}
	}
	for ns, latestChecks := range byNamespace {
		if _, ok := s[ns]; !ok {
			s[ns] = checkSet{}
		}
		prev := s[ns]
		a, c, r := prev.Update(latestChecks)
		s[ns] = prev
		added = append(added, a...)
		changed = append(changed, c...)
		removed = append(removed, r...)
	}

	return added, changed, removed
}

// checkSet is a collection of CheckConfigs
type checkSet struct {
	set []*corev2.CheckConfig
}

// Update the checkSet.
// returns a slice for CheckConfigs that have been added, changed and removed
func (sp *checkSet) Update(next []*corev2.CheckConfig) ([]*corev2.CheckConfig, []*corev2.CheckConfig, []*corev2.CheckConfig) {
	var added, changed, removed []*corev2.CheckConfig
	prev := sp.set

	sort.Slice(next, func(i, j int) bool { return next[i].Name < next[j].Name })

	p, n := prev, next
	for len(p) > 0 && len(n) > 0 {
		if p[0].Name == n[0].Name {
			if !p[0].Equal(n[0]) {
				changed = append(changed, n[0])
			}
			p, n = p[1:], n[1:]
			continue
		}

		if p[0].Name < n[0].Name {
			removed = append(removed, p[0])
			p = p[1:]
			continue
		}
		added = append(added, n[0])
		n = n[1:]
	}
	if len(p) > 0 {
		removed = append(removed, p...)
	}
	if len(n) > 0 {
		added = append(added, n...)
	}

	sp.set = next

	return added, changed, removed
}
