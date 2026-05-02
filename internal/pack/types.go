package pack

type Question struct {
	Value  int    `json:"value"`
	Text   string `json:"text"`
	Answer string `json:"answer"`
}

type Category struct {
	Name      string     `json:"name"`
	Questions []Question `json:"questions"`
}

type QuizPack struct {
	ID         string     `json:"id"`
	Lang       string     `json:"lang"`
	Title      string     `json:"title"`
	Categories []Category `json:"categories"`
}
