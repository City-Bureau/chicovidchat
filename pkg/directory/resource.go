package directory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Resource represents one item from the Airtable directory
type Resource struct {
	Name          string     `json:"Name"`
	Link          string     `json:"Link"`
	Phone         string     `json:"Phone"`
	Email         string     `json:"Email"`
	Description   string     `json:"Description"`
	DescriptionES string     `json:"Description ES"`
	Hours         string     `json:"Hours"`
	Languages     []string   `json:"Languages,omitempty"`
	Address       string     `json:"Address"`
	ZIP           string     `json:"ZIP"`
	Category      []string   `json:"Category,omitempty"`
	Who           []string   `json:"Who,omitempty"`
	Level         string     `json:"Level"`
	Type          string     `json:"Type"`
	Status        string     `json:"Status"`
	ExternalID    string     `json:"External ID"`
	LastUpdated   *time.Time `json:"Last Updated,omitempty"`
	Created       *time.Time `json:"Created,omitempty"`
}

// AsText should return a resource as it should display for a chat message
func (r *Resource) AsText(lang string, localizer *i18n.Localizer) string {
	resourceStr := r.Name
	if r.Category != nil && len(r.Category) > 0 {
		resourceStr += "\n"
		resourceStr += translateSlice(r.Category, localizer)
	}
	if r.Who != nil && len(r.Who) > 0 {
		resourceStr += "\n"
		resourceStr += translateSlice(r.Who, localizer)
	}
	if r.Languages != nil && len(r.Languages) > 0 {
		resourceStr += "\n"
		resourceStr += translateSlice(r.Languages, localizer)
	}

	langDescription := r.descriptionForLang(lang)
	if langDescription != "" {
		resourceStr += fmt.Sprintf("\n%s", langDescription)
	}
	if r.Hours != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Hours)
	}
	if r.Phone != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Phone)
	}
	if r.Address != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Address)
	}
	if r.Link != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Link)
	}
	if r.Email != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Email)
	}
	return resourceStr
}

func (r *Resource) descriptionForLang(lang string) string {
	switch {
	case lang == "es" && r.DescriptionES != "":
		return r.DescriptionES
	default:
		return r.Description
	}
}

func translateSlice(items []string, localizer *i18n.Localizer) string {
	translatedItems := []string{}
	for _, item := range items {
		translatedItems = append(translatedItems, localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: item, Other: item},
		}))
	}
	return strings.Join(translatedItems, ", ")
}

func (r *Resource) updateChangedFields(resource Resource) bool {
	changed := false
	if r.Name != resource.Name && resource.Name != "" {
		r.Name = resource.Name
		changed = true
	}
	if r.Link != resource.Link && resource.Link != "" {
		r.Link = resource.Link
		changed = true
	}
	if r.Description != resource.Description && resource.Description != "" {
		r.Description = resource.Description
		changed = true
	}
	if r.Hours != resource.Hours && resource.Hours != "" {
		r.Hours = resource.Hours
		changed = true
	}
	if r.Address != resource.Address && resource.Address != "" {
		r.Address = resource.Address
		changed = true
	}
	if r.ZIP != resource.ZIP && resource.ZIP != "" {
		r.ZIP = resource.ZIP
		changed = true
	}
	if slicesEqual(r.Category, resource.Category) && len(resource.Category) > 0 {
		r.Category = resource.Category
		changed = true
	}
	if slicesEqual(r.Who, resource.Who) && len(resource.Who) > 0 {
		r.Who = resource.Who
		changed = true
	}
	if r.Level != resource.Level && resource.Level != "" {
		r.Level = resource.Level
		changed = true
	}
	if r.Type != resource.Type && resource.Type != "" {
		r.Type = resource.Type
		changed = true
	}
	return changed
}

type airtableRecord struct {
	ID     string   `json:"id"`
	Fields Resource `json:"fields"`
}

type airtableResponse struct {
	Records []airtableRecord `json:"records"`
	Offset  *string          `json:"offset"`
}

func loadAirtableRecords(base, table, key string) ([]airtableRecord, error) {
	var records []airtableRecord
	offset := ""

	for {
		reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s?pageSize=100", base, table)
		log.Println(reqURL)
		if offset != "" {
			reqURL += fmt.Sprintf("&offset=%s", offset)
		}

		client := &http.Client{}
		req, _ := http.NewRequest("GET", reqURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
		res, err := client.Do(req)
		if err != nil {
			return records, err
		}
		defer res.Body.Close()

		body, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			return records, readErr
		}

		resMap := airtableResponse{}
		_ = json.Unmarshal(body, &resMap)

		airtableRes := airtableResponse{}
		jsonErr := json.Unmarshal(body, &airtableRes)
		if jsonErr != nil {
			return records, jsonErr
		}

		records = append(records, airtableRes.Records...)

		if airtableRes.Offset == nil {
			return records, nil
		}
		offset = *airtableRes.Offset
	}
}

// LoadAirtableResources loads the full resource directory table from Airtable
func LoadAirtableResources(base, table, key string) ([]Resource, error) {
	var resources []Resource

	records, err := loadAirtableRecords(base, table, key)
	if err != nil {
		return resources, err
	}

	for _, rec := range records {
		resources = append(resources, rec.Fields)
	}

	return resources, nil
}

// SyncAirtableResources takes a slice of resources and updates or creates them based on ExternalID
func SyncAirtableResources(resources []Resource, base, table, key string) error {
	recordIDMap := map[string]airtableRecord{}
	records, err := loadAirtableRecords(base, table, key)
	if err != nil {
		return err
	}

	for _, r := range records {
		if r.Fields.ExternalID != "" {
			recordIDMap[r.Fields.ExternalID] = r
		}
	}

	for _, r := range resources {
		// TODO: Make sure fields aren't dropped (modify?)
		if updateRec, ok := recordIDMap[r.ExternalID]; ok {
			changed := updateRec.Fields.updateChangedFields(r)
			if changed {
				updateErr := createOrUpdateAirtableResource(updateRec.Fields, &updateRec.ID, base, table, key)
				if updateErr != nil {
					return updateErr
				}
			}
		} else {
			createErr := createOrUpdateAirtableResource(r, nil, base, table, key)
			if createErr != nil {
				return createErr
			}
		}
	}
	log.Println(len(records))
	return nil
}

// LoadResources pulls the latest resource items from S3
func LoadResources() ([]Resource, error) {
	var resources []Resource
	sess, _ := session.NewSession()
	svc := s3.New(sess)

	results, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String("latest.json"),
	})
	if err != nil {
		return resources, err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return resources, err
	}

	jsonErr := json.Unmarshal(buf.Bytes(), &resources)
	if jsonErr != nil {
		return resources, jsonErr
	}

	return resources, nil
}

func createOrUpdateAirtableResource(resource Resource, id *string, base, table, key string) error {
	method := "POST"
	reqURL := fmt.Sprintf("https://api.airtable.com/v0/%s/%s", base, table)
	if id != nil {
		method = "PATCH"
		reqURL = fmt.Sprintf("%s/%s", reqURL, *id)
	}
	// TODO: Include body/fields
	client := &http.Client{}
	req, _ := http.NewRequest(method, reqURL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", key))
	_, err := client.Do(req)
	return err
}

func slicesEqual(a, b []string) bool {
	sort.Strings(a)
	sort.Strings(b)
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
