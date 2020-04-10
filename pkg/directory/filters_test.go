package directory

import (
	"testing"
)

func TestMatchesFiltersZIP(t *testing.T) {
	zipOne := "12345"
	zipTwo := "67890"
	zipThree := "45678"
	zipMap := &map[string][]string{
		zipOne: []string{zipOne, zipTwo},
	}
	params := FilterParams{}
	resource := Resource{ZIP: zipTwo, Status: "Approved"}
	if !params.MatchesFilters(resource, zipMap) {
		t.Errorf("Empty filters not matching")
	}
	params.ZIP = &zipThree
	if !params.MatchesFilters(resource, zipMap) {
		t.Errorf("Params not matching non-neighborhood on ZIP")
	}
	resource.Level = "Neighborhood"
	if params.MatchesFilters(resource, zipMap) {
		t.Errorf("Resource matching even though neighborhood")
	}
	params.ZIP = &zipOne
	resource.ZIP = zipTwo
	if !params.MatchesFilters(resource, zipMap) {
		t.Errorf("Resource not matching in map list")
	}
	params.ZIP = &zipTwo
	if !params.MatchesFilters(resource, zipMap) {
		t.Errorf("Resource ZIP not matching on strings")
	}
}

func TestMatchesFiltersWho(t *testing.T) {
	fakeZIP := "12345"
	params := FilterParams{ZIP: &fakeZIP}
	resourceNoWho := Resource{Who: []string{}, Status: "Approved"}
	resourceWho := Resource{Who: []string{"Families"}, Status: "Approved"}
	resourceWhoTwo := Resource{Who: []string{"Immigrants"}, Status: "Approved"}

	if !params.MatchesFilters(resourceNoWho, nil) || !params.MatchesFilters(resourceWho, nil) {
		t.Errorf("No who filters should match with and without who on resource")
	}

	params.Who = []string{"Families"}
	if !params.MatchesFilters(resourceNoWho, nil) || !params.MatchesFilters(resourceWho, nil) || params.MatchesFilters(resourceWhoTwo, nil) {
		t.Errorf("Filters with who not matching empty resources, matching, excluding other who resources")
	}
}

func TestMatchesFiltersWhoNone(t *testing.T) {
	fakeZIP := "12345"
	params := FilterParams{ZIP: &fakeZIP}
	resourceNoWho := Resource{Who: []string{}, Status: "Approved"}
	resourceWho := Resource{Who: []string{"Families"}, Status: "Approved"}
	resourceUntrackedWho := Resource{Who: []string{"TEST"}, Status: "Approved"}
	params.Who = []string{"None"}

	if !params.MatchesFilters(resourceNoWho, nil) {
		t.Errorf("Who filter for none should match resource with no who categories")
	}
	if params.MatchesFilters(resourceWho, nil) {
		t.Errorf("Who filter for none should not match resource with tracked who category")
	}
	if !params.MatchesFilters(resourceUntrackedWho, nil) {
		t.Errorf("Who filter for none should match resource with untracked who category")
	}
}
