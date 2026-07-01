package model

type AnswerMode string

const (
	AnswerModeManual AnswerMode = "manual"
	AnswerModeTiles  AnswerMode = "tiles"
)

type HintImage struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Label string `json:"label"`
	Alt   string `json:"alt"`
}

type CandidateChar struct {
	ID     string `json:"id"`
	Char   string `json:"char"`
	Pinyin string `json:"pinyin"`
}

type PuzzleSet struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	DomainType  string `json:"domain_type"`
	CoverURL    string `json:"cover_url"`
	PuzzleCount int    `json:"puzzle_count"`
}

type Puzzle struct {
	ID                   int64           `json:"id"`
	PuzzleSetID          int64           `json:"puzzle_set_id"`
	AuthorName           string          `json:"author_name"`
	HintImages           []HintImage     `json:"hint_images"`
	HintText             string          `json:"hint_text"`
	Answer               string          `json:"answer"`
	AnswerPinyin         string          `json:"answer_pinyin"`
	AnswerAliases        []string        `json:"answer_aliases"`
	AnswerLength         int             `json:"answer_length"`
	CandidateChars       []CandidateChar `json:"candidate_chars"`
	DefaultAnswerMode    AnswerMode      `json:"default_answer_mode"`
	SupportedAnswerModes []AnswerMode    `json:"supported_answer_modes"`
	BlankTemplate        string          `json:"blank_template"`
	Category             string          `json:"category"`
	Difficulty           int             `json:"difficulty"`
	Explanation          string          `json:"explanation"`
	SortOrder            int             `json:"sort_order"`
}

type PuzzlePublic struct {
	ID                   int64           `json:"id"`
	AttemptID            string          `json:"attempt_id,omitempty"`
	PuzzleSetID          int64           `json:"puzzle_set_id"`
	AuthorName           string          `json:"author_name"`
	HintImages           []HintImage     `json:"hint_images"`
	HintText             string          `json:"hint_text"`
	AnswerLength         int             `json:"answer_length"`
	CandidateChars       []CandidateChar `json:"candidate_chars"`
	DefaultAnswerMode    AnswerMode      `json:"default_answer_mode"`
	SupportedAnswerModes []AnswerMode    `json:"supported_answer_modes"`
	BlankTemplate        string          `json:"blank_template"`
	Category             string          `json:"category"`
	Difficulty           int             `json:"difficulty"`
	SortOrder            int             `json:"sort_order"`
}

func (p Puzzle) Public() PuzzlePublic {
	return PuzzlePublic{
		ID:                   p.ID,
		PuzzleSetID:          p.PuzzleSetID,
		AuthorName:           p.AuthorName,
		HintImages:           p.HintImages,
		HintText:             p.HintText,
		AnswerLength:         p.AnswerLength,
		CandidateChars:       p.CandidateChars,
		DefaultAnswerMode:    p.DefaultAnswerMode,
		SupportedAnswerModes: p.SupportedAnswerModes,
		BlankTemplate:        p.BlankTemplate,
		Category:             p.Category,
		Difficulty:           p.Difficulty,
		SortOrder:            p.SortOrder,
	}
}
