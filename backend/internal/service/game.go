package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mozillazg/go-pinyin"
	"golang.org/x/crypto/bcrypt"

	"this-is-pun/backend/internal/data"
	"this-is-pun/backend/internal/model"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidRequest = errors.New("invalid request")
	ErrAlreadyExists  = errors.New("already exists")
	ErrUnauthorized   = errors.New("unauthorized")
)

const adminSessionTTL = 12 * time.Hour

type Repository interface {
	ListPuzzleSets() []model.PuzzleSet
	GetPuzzleSet(id int64) (model.PuzzleSet, error)
	ListPuzzlesBySet(setID int64) ([]model.Puzzle, error)
	GetPuzzle(id int64) (model.Puzzle, error)
	ListAdminPuzzles() ([]model.Puzzle, error)
	CreatePuzzle(puzzle model.Puzzle) (model.Puzzle, error)
	UpdatePuzzle(id int64, puzzle model.Puzzle) (model.Puzzle, error)
	DeletePuzzle(id int64) error
	CreateAdminSession(session model.AdminSession) error
	GetAdminSession(token string) (model.AdminSession, error)
	DeleteAdminSession(token string) error
	CreateAdminUser(user model.AdminUser) (model.AdminUser, error)
	GetAdminUserByUsername(username string) (model.AdminUser, error)
	CreateSubmission(submission model.PuzzleSubmission) (model.PuzzleSubmission, error)
	ListSubmissions(status model.SubmissionStatus) ([]model.PuzzleSubmission, error)
	GetSubmission(id int64) (model.PuzzleSubmission, error)
	UpdateSubmissionStatus(id int64, status model.SubmissionStatus, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error)
	ApproveSubmission(id int64, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error)
}

type GameService struct {
	repo     Repository
	mu       sync.Mutex
	attempts map[string]attemptState
}

type attemptState struct {
	PuzzleID  int64
	HintsUsed int
}

type CheckAnswerRequest struct {
	AttemptID  string           `json:"attempt_id"`
	Answer     string           `json:"answer"`
	AnswerMode model.AnswerMode `json:"answer_mode"`
	ElapsedMS  int64            `json:"elapsed_ms"`
	HintsUsed  int              `json:"hints_used"`
}

type CheckAnswerResult struct {
	Correct       bool             `json:"correct"`
	Score         int              `json:"score"`
	Answer        string           `json:"answer,omitempty"`
	AnswerMode    model.AnswerMode `json:"answer_mode"`
	Normalized    string           `json:"normalized"`
	ExpectedChars int              `json:"expected_chars"`
	ElapsedMS     int64            `json:"elapsed_ms"`
	Explanation   string           `json:"explanation,omitempty"`
	Message       string           `json:"message"`
}

type HintRequest struct {
	AttemptID string `json:"attempt_id"`
	Level     int    `json:"level"`
}

type HintResult struct {
	Level          int    `json:"level"`
	Text           string `json:"text"`
	ScoreIfCorrect int    `json:"score_if_correct"`
}

type PuzzleInput struct {
	PuzzleSetID          int64                 `json:"puzzle_set_id"`
	AuthorName           string                `json:"author_name"`
	HintImages           []model.HintImage     `json:"hint_images"`
	HintText             string                `json:"hint_text"`
	Answer               string                `json:"answer"`
	AnswerPinyin         string                `json:"answer_pinyin"`
	AnswerAliases        []string              `json:"answer_aliases"`
	CandidateChars       []model.CandidateChar `json:"candidate_chars"`
	DefaultAnswerMode    model.AnswerMode      `json:"default_answer_mode"`
	SupportedAnswerModes []model.AnswerMode    `json:"supported_answer_modes"`
	BlankTemplate        string                `json:"blank_template"`
	Category             string                `json:"category"`
	Difficulty           int                   `json:"difficulty"`
	Explanation          string                `json:"explanation"`
	SortOrder            int                   `json:"sort_order"`
}

type SubmissionInput struct {
	CreatorName string `json:"creator_name"`
	Contact     string `json:"contact"`
	PuzzleInput
}

type AdminAuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminAuthResult struct {
	Token     string          `json:"token"`
	User      model.AdminUser `json:"user"`
	ExpiresAt time.Time       `json:"expires_at"`
}

type SubmissionReviewRequest struct {
	ReviewNote string `json:"review_note"`
}

func NewGameService(repo Repository) *GameService {
	return &GameService{repo: repo, attempts: map[string]attemptState{}}
}

func (s *GameService) ListPuzzleSets() []model.PuzzleSet {
	return s.repo.ListPuzzleSets()
}

func (s *GameService) GetPuzzleSet(id int64) (model.PuzzleSet, error) {
	return s.repo.GetPuzzleSet(id)
}

func (s *GameService) ListPublicPuzzlesBySet(setID int64) ([]model.PuzzlePublic, error) {
	puzzles, err := s.repo.ListPuzzlesBySet(setID)
	if err != nil {
		return nil, convertRepoError(err)
	}
	public := make([]model.PuzzlePublic, 0, len(puzzles))
	for _, puzzle := range puzzles {
		public = append(public, puzzle.Public())
	}
	return public, nil
}

func (s *GameService) GetPublicPuzzle(id int64) (model.PuzzlePublic, error) {
	puzzle, err := s.repo.GetPuzzle(id)
	if err != nil {
		return model.PuzzlePublic{}, convertRepoError(err)
	}
	public := puzzle.Public()
	public.AttemptID = s.createAttempt(id)
	return public, nil
}

func (s *GameService) CheckAnswer(id int64, req CheckAnswerRequest) (CheckAnswerResult, error) {
	puzzle, err := s.repo.GetPuzzle(id)
	if err != nil {
		return CheckAnswerResult{}, convertRepoError(err)
	}

	mode := req.AnswerMode
	if mode == "" {
		mode = puzzle.DefaultAnswerMode
	}
	if !supportsMode(puzzle, mode) {
		return CheckAnswerResult{}, ErrInvalidRequest
	}
	attempt, ok := s.getAttempt(req.AttemptID, id)
	if !ok {
		return CheckAnswerResult{}, ErrInvalidRequest
	}
	elapsedMS := req.ElapsedMS
	if elapsedMS < 0 {
		elapsedMS = 0
	}

	normalized := NormalizeAnswer(req.Answer)
	correct := s.isCorrect(puzzle, normalized)
	result := CheckAnswerResult{
		Correct:       correct,
		Score:         scoreForHints(attempt.HintsUsed),
		AnswerMode:    mode,
		Normalized:    normalized,
		ExpectedChars: puzzle.AnswerLength,
		ElapsedMS:     elapsedMS,
	}

	if correct {
		result.Answer = puzzle.Answer
		result.Explanation = puzzle.Explanation
		result.Message = "答对啦"
	} else {
		result.Score = 0
		result.Message = "还差一点，再调整一下"
	}

	return result, nil
}

func (s *GameService) GetHint(id int64, req HintRequest) (HintResult, error) {
	puzzle, err := s.repo.GetPuzzle(id)
	if err != nil {
		return HintResult{}, convertRepoError(err)
	}
	if _, ok := s.getAttempt(req.AttemptID, id); !ok {
		return HintResult{}, ErrInvalidRequest
	}
	level := clampHints(req.Level)
	if level < 1 {
		level = 1
	}
	s.recordHint(req.AttemptID, level)

	text := ""
	switch level {
	case 1:
		text = "分类：" + puzzle.Category + "，答案 " + strconv.Itoa(puzzle.AnswerLength) + " 个字"
	case 2:
		text = "第一个字：" + firstRune(puzzle.Answer)
	case 3:
		text = "拼音：" + puzzle.AnswerPinyin
	}

	return HintResult{
		Level:          level,
		Text:           text,
		ScoreIfCorrect: scoreForHints(level),
	}, nil
}

func (s *GameService) GetExplanation(id int64) (map[string]string, error) {
	puzzle, err := s.repo.GetPuzzle(id)
	if err != nil {
		return nil, convertRepoError(err)
	}
	return map[string]string{
		"answer":      puzzle.Answer,
		"explanation": puzzle.Explanation,
	}, nil
}

func (s *GameService) ListAdminPuzzles() ([]model.Puzzle, error) {
	return s.repo.ListAdminPuzzles()
}

func (s *GameService) CreatePuzzle(input PuzzleInput) (model.Puzzle, error) {
	puzzle, err := s.preparePuzzle(input)
	if err != nil {
		return model.Puzzle{}, err
	}
	return s.repo.CreatePuzzle(puzzle)
}

func (s *GameService) UpdatePuzzle(id int64, input PuzzleInput) (model.Puzzle, error) {
	puzzle, err := s.preparePuzzle(input)
	if err != nil {
		return model.Puzzle{}, err
	}
	return s.repo.UpdatePuzzle(id, puzzle)
}

func (s *GameService) DeletePuzzle(id int64) error {
	return s.repo.DeletePuzzle(id)
}

func (s *GameService) RegisterAdmin(req AdminAuthRequest) (AdminAuthResult, error) {
	username := normalizeUsername(req.Username)
	if username == "" || len(req.Password) < 6 {
		return AdminAuthResult{}, ErrInvalidRequest
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AdminAuthResult{}, err
	}
	user, err := s.repo.CreateAdminUser(model.AdminUser{
		Username:     username,
		PasswordHash: string(hash),
	})
	if err != nil {
		return AdminAuthResult{}, convertRepoError(err)
	}
	return s.issueAdminToken(user)
}

func (s *GameService) LoginAdmin(req AdminAuthRequest) (AdminAuthResult, error) {
	username := normalizeUsername(req.Username)
	if username == "" || req.Password == "" {
		return AdminAuthResult{}, ErrInvalidRequest
	}
	user, err := s.repo.GetAdminUserByUsername(username)
	if err != nil {
		return AdminAuthResult{}, ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AdminAuthResult{}, ErrUnauthorized
	}
	return s.issueAdminToken(user)
}

func (s *GameService) AdminSession(token string) (model.AdminSession, bool) {
	session, err := s.repo.GetAdminSession(token)
	if err != nil {
		return model.AdminSession{}, false
	}
	if time.Now().After(session.ExpiresAt) {
		_ = s.repo.DeleteAdminSession(token)
		return model.AdminSession{}, false
	}
	return session, true
}

func (s *GameService) CreateSubmission(input SubmissionInput) (model.PuzzleSubmission, error) {
	puzzle, err := s.preparePuzzle(input.PuzzleInput)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	creatorName := strings.TrimSpace(input.CreatorName)
	if creatorName == "" {
		creatorName = "匿名玩家"
	}
	submission := model.PuzzleSubmission{
		CreatorName:          creatorName,
		Contact:              strings.TrimSpace(input.Contact),
		Status:               model.SubmissionStatusPending,
		PuzzleSetID:          puzzle.PuzzleSetID,
		AuthorName:           puzzle.AuthorName,
		HintImages:           puzzle.HintImages,
		HintText:             puzzle.HintText,
		Answer:               puzzle.Answer,
		AnswerPinyin:         puzzle.AnswerPinyin,
		AnswerAliases:        puzzle.AnswerAliases,
		AnswerLength:         puzzle.AnswerLength,
		CandidateChars:       puzzle.CandidateChars,
		DefaultAnswerMode:    puzzle.DefaultAnswerMode,
		SupportedAnswerModes: puzzle.SupportedAnswerModes,
		BlankTemplate:        puzzle.BlankTemplate,
		Category:             puzzle.Category,
		Difficulty:           puzzle.Difficulty,
		Explanation:          puzzle.Explanation,
		SortOrder:            puzzle.SortOrder,
	}
	return s.repo.CreateSubmission(submission)
}

func (s *GameService) ListSubmissions(status model.SubmissionStatus) ([]model.PuzzleSubmission, error) {
	if status != "" && !isValidSubmissionStatus(status) {
		return nil, ErrInvalidRequest
	}
	return s.repo.ListSubmissions(status)
}

func (s *GameService) ApproveSubmission(id int64, reviewer string, req SubmissionReviewRequest) (model.PuzzleSubmission, error) {
	submission, err := s.repo.ApproveSubmission(id, strings.TrimSpace(req.ReviewNote), reviewer)
	if err != nil {
		return model.PuzzleSubmission{}, convertRepoError(err)
	}
	return submission, nil
}

func (s *GameService) RejectSubmission(id int64, reviewer string, req SubmissionReviewRequest) (model.PuzzleSubmission, error) {
	submission, err := s.repo.GetSubmission(id)
	if err != nil {
		return model.PuzzleSubmission{}, convertRepoError(err)
	}
	if submission.Status != model.SubmissionStatusPending {
		return model.PuzzleSubmission{}, ErrInvalidRequest
	}
	submission, err = s.repo.UpdateSubmissionStatus(id, model.SubmissionStatusRejected, strings.TrimSpace(req.ReviewNote), reviewer)
	if err != nil {
		return model.PuzzleSubmission{}, convertRepoError(err)
	}
	return submission, nil
}

func (s *GameService) preparePuzzle(input PuzzleInput) (model.Puzzle, error) {
	answer := strings.TrimSpace(input.Answer)
	if answer == "" {
		return model.Puzzle{}, ErrInvalidRequest
	}
	puzzleSetID := input.PuzzleSetID
	if puzzleSetID == 0 {
		puzzleSetID = 1
	}
	if _, err := s.repo.GetPuzzleSet(puzzleSetID); err != nil {
		return model.Puzzle{}, convertRepoError(err)
	}

	answerLength := RuneCount(answer)
	answerPinyin := strings.TrimSpace(input.AnswerPinyin)
	if answerPinyin == "" {
		answerPinyin = ToPinyin(answer)
	}
	blankTemplate := strings.TrimSpace(input.BlankTemplate)
	if blankTemplate == "" {
		blankTemplate = "这是" + strings.Repeat("_", answerLength)
	}

	difficulty := input.Difficulty
	if difficulty <= 0 {
		difficulty = 1
	}
	if difficulty > 5 {
		difficulty = 5
	}

	defaultMode := input.DefaultAnswerMode
	if defaultMode == "" {
		defaultMode = model.AnswerModeManual
	}
	supportedModes := normalizeModes(input.SupportedAnswerModes, defaultMode)
	if !isValidMode(defaultMode) || !modeInList(defaultMode, supportedModes) {
		return model.Puzzle{}, ErrInvalidRequest
	}

	authorName := strings.TrimSpace(input.AuthorName)
	if authorName == "" {
		authorName = "QQ"
	}

	return model.Puzzle{
		PuzzleSetID:          puzzleSetID,
		AuthorName:           authorName,
		HintImages:           normalizeHintImages(input.HintImages),
		HintText:             strings.TrimSpace(input.HintText),
		Answer:               answer,
		AnswerPinyin:         answerPinyin,
		AnswerAliases:        normalizeAliases(input.AnswerAliases),
		AnswerLength:         answerLength,
		CandidateChars:       normalizeCandidateChars(input.CandidateChars, answer),
		DefaultAnswerMode:    defaultMode,
		SupportedAnswerModes: supportedModes,
		BlankTemplate:        blankTemplate,
		Category:             strings.TrimSpace(input.Category),
		Difficulty:           difficulty,
		Explanation:          strings.TrimSpace(input.Explanation),
		SortOrder:            input.SortOrder,
	}, nil
}

func (s *GameService) isCorrect(puzzle model.Puzzle, normalized string) bool {
	if normalized == "" {
		return false
	}
	if normalized == NormalizeAnswer(puzzle.Answer) {
		return true
	}
	for _, alias := range puzzle.AnswerAliases {
		if normalized == NormalizeAnswer(alias) {
			return true
		}
	}
	return ToPinyin(normalized) == puzzle.AnswerPinyin
}

func NormalizeAnswer(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\n", "",
		"，", "",
		"。", "",
		"！", "",
		"？", "",
		",", "",
		".", "",
		"!", "",
		"?", "",
		"＿", "_",
	)
	input = replacer.Replace(input)
	fullWidthDigits := regexp.MustCompile(`[０-９]`)
	return fullWidthDigits.ReplaceAllStringFunc(input, func(s string) string {
		r := []rune(s)[0]
		return string(r - '０' + '0')
	})
}

func ToPinyin(input string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	parts := pinyin.Pinyin(input, args)
	words := make([]string, 0, len(parts))
	for _, item := range parts {
		if len(item) > 0 {
			words = append(words, item[0])
		}
	}
	return strings.Join(words, " ")
}

func scoreForHints(hintsUsed int) int {
	hintsUsed = clampHints(hintsUsed)
	switch {
	case hintsUsed <= 0:
		return 100
	case hintsUsed == 1:
		return 70
	case hintsUsed == 2:
		return 40
	default:
		return 20
	}
}

func clampHints(hintsUsed int) int {
	if hintsUsed < 0 {
		return 0
	}
	if hintsUsed > 3 {
		return 3
	}
	return hintsUsed
}

func supportsMode(puzzle model.Puzzle, mode model.AnswerMode) bool {
	for _, supported := range puzzle.SupportedAnswerModes {
		if mode == supported {
			return true
		}
	}
	return false
}

func (s *GameService) issueAdminToken(user model.AdminUser) (AdminAuthResult, error) {
	token := randomID()
	expiresAt := time.Now().Add(adminSessionTTL)
	session := model.AdminSession{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.CreateAdminSession(session); err != nil {
		return AdminAuthResult{}, err
	}
	return AdminAuthResult{Token: token, User: user, ExpiresAt: expiresAt}, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func isValidSubmissionStatus(status model.SubmissionStatus) bool {
	return status == model.SubmissionStatusPending || status == model.SubmissionStatusApproved || status == model.SubmissionStatusRejected
}

func normalizeModes(modes []model.AnswerMode, defaultMode model.AnswerMode) []model.AnswerMode {
	normalized := make([]model.AnswerMode, 0, len(modes)+1)
	for _, mode := range modes {
		if isValidMode(mode) && !modeInList(mode, normalized) {
			normalized = append(normalized, mode)
		}
	}
	if len(normalized) == 0 {
		normalized = []model.AnswerMode{model.AnswerModeManual, model.AnswerModeTiles}
	}
	if isValidMode(defaultMode) && !modeInList(defaultMode, normalized) {
		normalized = append(normalized, defaultMode)
	}
	return normalized
}

func isValidMode(mode model.AnswerMode) bool {
	return mode == model.AnswerModeManual || mode == model.AnswerModeTiles
}

func modeInList(mode model.AnswerMode, modes []model.AnswerMode) bool {
	for _, item := range modes {
		if item == mode {
			return true
		}
	}
	return false
}

func normalizeAliases(aliases []string) []string {
	normalized := make([]string, 0, len(aliases))
	seen := map[string]bool{}
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		key := NormalizeAnswer(alias)
		if alias == "" || seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, alias)
	}
	return normalized
}

func normalizeHintImages(images []model.HintImage) []model.HintImage {
	normalized := make([]model.HintImage, 0, len(images))
	for index, image := range images {
		if image.ID == "" {
			image.ID = "hint-" + strconv.Itoa(index+1)
		}
		if strings.TrimSpace(image.URL) == "" {
			image.URL = "emoji:❓"
		}
		if strings.TrimSpace(image.Label) == "" {
			image.Label = "提示图 " + strconv.Itoa(index+1)
		}
		if strings.TrimSpace(image.Alt) == "" {
			image.Alt = image.Label
		}
		normalized = append(normalized, image)
	}
	for len(normalized) < 2 {
		index := len(normalized) + 1
		normalized = append(normalized, model.HintImage{
			ID:    "hint-" + strconv.Itoa(index),
			URL:   "emoji:❓",
			Label: "提示图 " + strconv.Itoa(index),
			Alt:   "提示图 " + strconv.Itoa(index),
		})
	}
	return normalized
}

func normalizeCandidateChars(candidates []model.CandidateChar, answer string) []model.CandidateChar {
	if len(candidates) == 0 {
		for _, char := range answer {
			candidates = append(candidates, model.CandidateChar{Char: string(char)})
		}
	}

	normalized := make([]model.CandidateChar, 0, len(candidates))
	for index, candidate := range candidates {
		candidate.Char = strings.TrimSpace(candidate.Char)
		if candidate.Char == "" {
			continue
		}
		if candidate.ID == "" {
			candidate.ID = "c" + strconv.Itoa(index+1)
		}
		if strings.TrimSpace(candidate.Pinyin) == "" {
			candidate.Pinyin = ToPinyin(candidate.Char)
		}
		normalized = append(normalized, candidate)
	}
	return normalized
}

func (s *GameService) createAttempt(puzzleID int64) string {
	id := randomID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts[id] = attemptState{PuzzleID: puzzleID}
	return id
}

func (s *GameService) getAttempt(id string, puzzleID int64) (attemptState, bool) {
	if id == "" {
		return attemptState{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt, ok := s.attempts[id]
	return attempt, ok && attempt.PuzzleID == puzzleID
}

func (s *GameService) recordHint(id string, level int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	attempt := s.attempts[id]
	if level > attempt.HintsUsed {
		attempt.HintsUsed = level
		s.attempts[id] = attempt
	}
}

func randomID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(bytes[:])
}

func firstRune(input string) string {
	for _, r := range input {
		return string(r)
	}
	return ""
}

func convertRepoError(err error) error {
	if errors.Is(err, data.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, data.ErrAlreadyExists) {
		return ErrAlreadyExists
	}
	if errors.Is(err, data.ErrInvalidState) {
		return ErrInvalidRequest
	}
	return err
}

func RuneCount(input string) int {
	count := 0
	for _, r := range input {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}
