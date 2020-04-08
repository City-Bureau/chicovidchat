package directory

import (
	"strings"
)

// FilterParams represent all categories a user can filter in a chat by
type FilterParams struct {
	What      []string `json:"what"`
	Who       []string `json:"who"`
	Languages []string `json:"languages"`
	ZIP       *string  `json:"zip"`
}

func (f *FilterParams) isEmpty() bool {
	return len(f.What) == 0 && len(f.Who) == 0 && len(f.Languages) == 0 && f.ZIP == nil
}

// MatchesFilters determines whether a resource matches filter parameters
func (f *FilterParams) MatchesFilters(resource Resource, zipMap *map[string][]string) bool {
	// If filters are empty, return true
	if f.isEmpty() {
		return true
	}
	if resource.Status != "Approved" {
		return false
	}
	zipMatches := false
	if f.ZIP != nil {
		if resource.Level != "Neighborhood" {
			zipMatches = true
		} else if zipMap != nil {
			if zipMatchList, ok := (*zipMap)[*f.ZIP]; ok {
				zipMatches = stringSlicesOverlap([]string{*f.ZIP}, zipMatchList)
			} else {
				zipMatches = strings.Contains(resource.ZIP, *f.ZIP)
			}
		} else {
			zipMatches = strings.Contains(resource.ZIP, *f.ZIP)
		}
	}

	whatMatches := len(f.What) == 0 || stringSlicesOverlap(f.What, resource.Category)
	// Includes unrestricted resources in addition to filtered ones
	whoMatches := len(f.Who) == 0 || len(resource.Who) == 0 || stringSlicesOverlap(f.Who, resource.Who)
	langMatches := len(f.Languages) == 0 || stringSlicesOverlap(f.Languages, resource.Languages)
	return whatMatches && whoMatches && langMatches && zipMatches
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
