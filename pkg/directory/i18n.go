package directory

import (
	"encoding/json"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func LoadLocalizer(lang string) *i18n.Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.MustLoadMessageFile("i18n/en.json")
	bundle.MustLoadMessageFile(fmt.Sprintf("i18n/%s.json", lang))
	if lang != "" {
		return i18n.NewLocalizer(bundle, lang, "en")
	}
	return i18n.NewLocalizer(bundle, "en")
}
