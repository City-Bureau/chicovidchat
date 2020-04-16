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
	if !params.MatchesFilters(resource, zipMap, nil) {
		t.Errorf("Empty filters not matching")
	}
	params.ZIP = &zipThree
	if !params.MatchesFilters(resource, zipMap, nil) {
		t.Errorf("Params not matching non-neighborhood on ZIP")
	}
	resource.Level = "Neighborhood"
	if params.MatchesFilters(resource, zipMap, nil) {
		t.Errorf("Resource matching even though neighborhood")
	}
	params.ZIP = &zipOne
	resource.ZIP = zipTwo
	if !params.MatchesFilters(resource, zipMap, nil) {
		t.Errorf("Resource not matching in map list")
	}
	params.ZIP = &zipTwo
	if !params.MatchesFilters(resource, zipMap, nil) {
		t.Errorf("Resource ZIP not matching on strings")
	}
}

func TestMatchesFiltersWho(t *testing.T) {
	fakeZIP := "12345"
	params := FilterParams{ZIP: &fakeZIP}
	resourceNoWho := Resource{Who: []string{}, Status: "Approved"}
	resourceWho := Resource{Who: []string{"Families"}, Status: "Approved"}
	resourceWhoTwo := Resource{Who: []string{"Immigrants"}, Status: "Approved"}

	if !params.MatchesFilters(resourceNoWho, nil, nil) || !params.MatchesFilters(resourceWho, nil, nil) {
		t.Errorf("No who filters should match with and without who on resource")
	}

	params.Who = []string{"Families"}
	if !params.MatchesFilters(resourceNoWho, nil, nil) || !params.MatchesFilters(resourceWho, nil, nil) || params.MatchesFilters(resourceWhoTwo, nil, nil) {
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

	if !params.MatchesFilters(resourceNoWho, nil, nil) {
		t.Errorf("Who filter for none should match resource with no who categories")
	}
	if params.MatchesFilters(resourceWho, nil, nil) {
		t.Errorf("Who filter for none should not match resource with tracked who category")
	}
	if !params.MatchesFilters(resourceUntrackedWho, nil, nil) {
		t.Errorf("Who filter for none should match resource with untracked who category")
	}
}

func TestMatchesFiltersCity(t *testing.T) {
	fakeZip := "12345"
	cityZip := "23456"
	cityZips := []string{cityZip}

	params := FilterParams{ZIP: &fakeZip}
	resource := Resource{ZIP: "34567", Level: "City", Status: "Approved"}
	if !params.MatchesFilters(resource, nil, nil) {
		t.Errorf("Empty chiZips not matching City")
	}

	if params.MatchesFilters(resource, nil, &cityZips) {
		t.Errorf("City resource should not match if cityZips does not include ZIP")
	}

	cityZips = []string{fakeZip, cityZip}
	if !params.MatchesFilters(resource, nil, &cityZips) {
		t.Errorf("City resource should match if filter ZIP in cityZips")
	}
}
