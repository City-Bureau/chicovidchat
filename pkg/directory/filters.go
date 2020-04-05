package directory

import "strings"

// FilterParams represent all categories a user can filter in a chat by
type FilterParams struct {
	What      []string `json:"what"`
	Who       []string `json:"who"`
	Languages []string `json:"Languages"`
	ZIP       *string  `json:"zip"`
}

func (f *FilterParams) isEmpty() bool {
	return len(f.What) == 0 && len(f.Who) == 0 && len(f.Languages) == 0 && f.ZIP == nil
}

// MatchesFilters determines whether a resource matches filter parameters
func (f *FilterParams) MatchesFilters(resource Resource) bool {
	// If filters are empty, return true
	if f.isEmpty() {
		return true
	}

	// TODO: Figure out logic here
	return (stringSlicesOverlap(f.What, resource.Category)) ||
		// TODO: Who should include non-restricted as well as specifically filtered for to reduce redoing
		(stringSlicesOverlap(f.Who, resource.Who)) ||
		(stringSlicesOverlap(f.Languages, resource.Languages)) ||
		(f.ZIP != nil && strings.Contains(resource.ZIP, (*f.ZIP)))

}

func stringSlicesOverlap(sliceA []string, sliceB []string) bool {
	for _, a := range sliceA {
		for _, b := range sliceB {
			if a == b {
				return true
			}
		}
	}
	return false
}
