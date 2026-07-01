package model

import "time"

type SubmissionStatus string

const (
	SubmissionStatusPending  SubmissionStatus = "pending"
	SubmissionStatusApproved SubmissionStatus = "approved"
	SubmissionStatusRejected SubmissionStatus = "rejected"
)

type AdminUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type AdminSession struct {
	Token     string    `json:"-"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type PuzzleSubmission struct {
	ID                   int64            `json:"id"`
	CreatorName          string           `json:"creator_name"`
	Contact              string           `json:"contact"`
	Status               SubmissionStatus `json:"status"`
	ReviewNote           string           `json:"review_note"`
	ReviewedBy           string           `json:"reviewed_by"`
	PuzzleSetID          int64            `json:"puzzle_set_id"`
	AuthorName           string           `json:"author_name"`
	HintImages           []HintImage      `json:"hint_images"`
	HintText             string           `json:"hint_text"`
	Answer               string           `json:"answer"`
	AnswerPinyin         string           `json:"answer_pinyin"`
	AnswerAliases        []string         `json:"answer_aliases"`
	AnswerLength         int              `json:"answer_length"`
	CandidateChars       []CandidateChar  `json:"candidate_chars"`
	DefaultAnswerMode    AnswerMode       `json:"default_answer_mode"`
	SupportedAnswerModes []AnswerMode     `json:"supported_answer_modes"`
	BlankTemplate        string           `json:"blank_template"`
	Category             string           `json:"category"`
	Difficulty           int              `json:"difficulty"`
	Explanation          string           `json:"explanation"`
	SortOrder            int              `json:"sort_order"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
	ApprovedPuzzle       *PuzzlePublic    `json:"approved_puzzle,omitempty"`
}

func (s PuzzleSubmission) ToPuzzle() Puzzle {
	return Puzzle{
		PuzzleSetID:          s.PuzzleSetID,
		AuthorName:           s.AuthorName,
		HintImages:           s.HintImages,
		HintText:             s.HintText,
		Answer:               s.Answer,
		AnswerPinyin:         s.AnswerPinyin,
		AnswerAliases:        s.AnswerAliases,
		AnswerLength:         s.AnswerLength,
		CandidateChars:       s.CandidateChars,
		DefaultAnswerMode:    s.DefaultAnswerMode,
		SupportedAnswerModes: s.SupportedAnswerModes,
		BlankTemplate:        s.BlankTemplate,
		Category:             s.Category,
		Difficulty:           s.Difficulty,
		Explanation:          s.Explanation,
		SortOrder:            s.SortOrder,
	}
}
