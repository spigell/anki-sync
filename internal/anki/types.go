package anki

type Data struct {
	Models []Model
	Decks  []Deck
}

type Model struct {
	Name          string         `yaml:"name" json:"modelName"`
	InOrderFields []string       `yaml:"fields" json:"inOrderFields"`
	CSS           string         `yaml:"css,omitempty" json:"css,omitempty"`
	IsCloze       bool           `yaml:"isCloze,omitempty" json:"isCloze,omitempty"`
	CardTemplates []CardTemplate `yaml:"cardTemplates" json:"cardTemplates"`
}

type CardTemplate struct {
	Name  string `yaml:"name" json:"Name"`
	Front string `yaml:"front" json:"Front"`
	Back  string `yaml:"back" json:"Back"`
}

type Deck struct {
	Deck         string `yaml:"deck_name"`
	Model        string `yaml:"model_name"`
	PrimaryField string `yaml:"primary_field"`
	Notes        []Note
}

type Note struct {
	Fields map[string]string `yaml:"fields"`
	Tags   []string          `yaml:"tags"`
}
