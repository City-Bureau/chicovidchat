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
	resourceStr := fmt.Sprintf("%s\n", r.Name)
	if r.Category != nil && len(r.Category) > 0 {
		resourceStr += fmt.Sprintf("\n%s: %s", localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "what-label",
		}), translateSlice(r.Category, localizer))
	}
	if r.Who != nil && len(r.Who) > 0 {
		resourceStr += fmt.Sprintf("\n%s: %s", localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "who-label",
		}), translateSlice(r.Who, localizer))
	}
	if r.Languages != nil && len(r.Languages) > 0 {
		resourceStr += fmt.Sprintf("\n%s: %s", localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "languages-label",
		}), translateSlice(r.Languages, localizer))
	}
	if r.Hours != "" {
		resourceStr += fmt.Sprintf("\n%s: %s", localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "hours-label",
		}), r.Hours)
	}

	langDescription := r.descriptionForLang(lang)
	if langDescription != "" {
		resourceStr += fmt.Sprintf("\n\n%s\n", strings.TrimSpace(langDescription))
	}
	if r.Phone != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Phone)
	}
	if r.Link != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Link)
	}
	if r.Email != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Email)
	}
	if r.Address != "" {
		resourceStr += fmt.Sprintf("\n%s", r.Address)
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

	levelOrder := map[string]int{
		"National":     0,
		"State":        1,
		"County":       2,
		"City":         3,
		"Neighborhood": 4,
	}
	sort.SliceStable(resources, func(a, b int) bool {
		aVal, aOk := levelOrder[resources[a].Level]
		if !aOk {
			aVal = 10
		}
		bVal, bOk := levelOrder[resources[b].Level]
		if !bOk {
			bVal = 10
		}
		return aVal < bVal
	})

	return resources, nil
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
