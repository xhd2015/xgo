package lang

type ProfileLanguage string

const (
	ProfileLanguage_Default ProfileLanguage = ""
	ProfileLanguage_Go      ProfileLanguage = "go"
	ProfileLanguage_Js      ProfileLanguage = "js"
)

func (c ProfileLanguage) IsGo() bool {
	return c == ProfileLanguage_Default || c == ProfileLanguage_Go
}
