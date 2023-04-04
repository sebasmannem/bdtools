package quotagroups

import (
	"golang.org/x/exp/maps"
)

type groups map[string]bool

func (og groups) List() []string {
	var l []string
	for key := range og {
		l = append(l, key)
	}
	return l
}

func (og groups) Intersection(other groups) groups {
	log.Debug("Intersecting %v, with %v", og.List(), other.List())
	both := make(groups)
	for key := range other {
		if _, exists := og[key]; exists {
			both[key] = true
		}
	}
	log.Debug("Intersection: %e", both.List())
	return both
}

func (og groups) Difference(other groups) groups {
	log.Debug("Difference between %v, with %v", og.List(), other.List())
	lonely := make(groups)
	for key := range og {
		if _, exists := other[key]; !exists {
			lonely[key] = true
		}
	}
	log.Debug("Difference: %e", lonely.List())
	return lonely
}

func (og groups) Combine(other groups) groups {
	log.Debug("Combining %v with %v", og.List(), other.List())
	combined := make(groups)
	maps.Copy(combined, og)
	maps.Copy(combined, other)
	log.Debug("Combination: %e", combined.List())
	return combined
}
