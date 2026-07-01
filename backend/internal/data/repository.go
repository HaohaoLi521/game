package data

import (
	"errors"
	"sort"
	"sync"
	"time"

	"this-is-pun/backend/internal/model"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidState  = errors.New("invalid state")
)

type SeedData struct {
	Sets    []model.PuzzleSet
	Puzzles []model.Puzzle
}

type MemoryRepository struct {
	mu           sync.RWMutex
	sets         map[int64]model.PuzzleSet
	puzzles      map[int64]model.Puzzle
	puzzlesBySet map[int64][]model.Puzzle
	adminUsers   map[string]model.AdminUser
	sessions     map[string]model.AdminSession
	submissions  map[int64]model.PuzzleSubmission
}

func NewMemoryRepository(seed SeedData) *MemoryRepository {
	repo := &MemoryRepository{
		sets:         map[int64]model.PuzzleSet{},
		puzzles:      map[int64]model.Puzzle{},
		puzzlesBySet: map[int64][]model.Puzzle{},
		adminUsers:   map[string]model.AdminUser{},
		sessions:     map[string]model.AdminSession{},
		submissions:  map[int64]model.PuzzleSubmission{},
	}

	for _, set := range seed.Sets {
		repo.sets[set.ID] = set
	}
	for _, puzzle := range seed.Puzzles {
		repo.puzzles[puzzle.ID] = puzzle
		repo.puzzlesBySet[puzzle.PuzzleSetID] = append(repo.puzzlesBySet[puzzle.PuzzleSetID], puzzle)
	}
	for setID, puzzles := range repo.puzzlesBySet {
		sort.Slice(puzzles, func(i, j int) bool {
			return puzzles[i].SortOrder < puzzles[j].SortOrder
		})
		repo.puzzlesBySet[setID] = puzzles
	}

	return repo
}

func (r *MemoryRepository) ListPuzzleSets() []model.PuzzleSet {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sets := make([]model.PuzzleSet, 0, len(r.sets))
	for _, set := range r.sets {
		set.PuzzleCount = len(r.puzzlesBySet[set.ID])
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool {
		return sets[i].ID < sets[j].ID
	})
	return sets
}

func (r *MemoryRepository) GetPuzzleSet(id int64) (model.PuzzleSet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	set, ok := r.sets[id]
	if !ok {
		return model.PuzzleSet{}, ErrNotFound
	}
	set.PuzzleCount = len(r.puzzlesBySet[id])
	return set, nil
}

func (r *MemoryRepository) ListPuzzlesBySet(setID int64) ([]model.Puzzle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.sets[setID]; !ok {
		return nil, ErrNotFound
	}
	return append([]model.Puzzle(nil), r.puzzlesBySet[setID]...), nil
}

func (r *MemoryRepository) GetPuzzle(id int64) (model.Puzzle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	puzzle, ok := r.puzzles[id]
	if !ok {
		return model.Puzzle{}, ErrNotFound
	}
	return puzzle, nil
}

func (r *MemoryRepository) ListAdminPuzzles() ([]model.Puzzle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	puzzles := make([]model.Puzzle, 0, len(r.puzzles))
	for _, puzzle := range r.puzzles {
		puzzles = append(puzzles, puzzle)
	}
	sortPuzzles(puzzles)
	return puzzles, nil
}

func (r *MemoryRepository) CreatePuzzle(puzzle model.Puzzle) (model.Puzzle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.sets[puzzle.PuzzleSetID]; !ok {
		return model.Puzzle{}, ErrNotFound
	}
	if puzzle.ID == 0 {
		puzzle.ID = r.nextPuzzleID()
	}
	r.puzzles[puzzle.ID] = puzzle
	r.rebuildPuzzlesBySet()
	return puzzle, nil
}

func (r *MemoryRepository) UpdatePuzzle(id int64, puzzle model.Puzzle) (model.Puzzle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.puzzles[id]; !ok {
		return model.Puzzle{}, ErrNotFound
	}
	if _, ok := r.sets[puzzle.PuzzleSetID]; !ok {
		return model.Puzzle{}, ErrNotFound
	}
	puzzle.ID = id
	r.puzzles[id] = puzzle
	r.rebuildPuzzlesBySet()
	return puzzle, nil
}

func (r *MemoryRepository) DeletePuzzle(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.puzzles[id]; !ok {
		return ErrNotFound
	}
	delete(r.puzzles, id)
	r.rebuildPuzzlesBySet()
	return nil
}

func (r *MemoryRepository) CreateAdminSession(session model.AdminSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session.CreatedAt = time.Now()
	r.sessions[session.Token] = session
	return nil
}

func (r *MemoryRepository) GetAdminSession(token string) (model.AdminSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[token]
	if !ok {
		return model.AdminSession{}, ErrNotFound
	}
	return session, nil
}

func (r *MemoryRepository) DeleteAdminSession(token string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.sessions, token)
	return nil
}

func (r *MemoryRepository) CreateAdminUser(user model.AdminUser) (model.AdminUser, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.adminUsers[user.Username]; ok {
		return model.AdminUser{}, ErrAlreadyExists
	}
	user.ID = int64(len(r.adminUsers) + 1)
	user.CreatedAt = time.Now()
	r.adminUsers[user.Username] = user
	return user, nil
}

func (r *MemoryRepository) GetAdminUserByUsername(username string) (model.AdminUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.adminUsers[username]
	if !ok {
		return model.AdminUser{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryRepository) CreateSubmission(submission model.PuzzleSubmission) (model.PuzzleSubmission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	submission.ID = r.nextSubmissionID()
	submission.Status = model.SubmissionStatusPending
	now := time.Now()
	submission.CreatedAt = now
	submission.UpdatedAt = now
	r.submissions[submission.ID] = submission
	return submission, nil
}

func (r *MemoryRepository) ListSubmissions(status model.SubmissionStatus) ([]model.PuzzleSubmission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	submissions := make([]model.PuzzleSubmission, 0, len(r.submissions))
	for _, submission := range r.submissions {
		if status == "" || submission.Status == status {
			submissions = append(submissions, submission)
		}
	}
	sortSubmissions(submissions)
	return submissions, nil
}

func (r *MemoryRepository) GetSubmission(id int64) (model.PuzzleSubmission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	submission, ok := r.submissions[id]
	if !ok {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	return submission, nil
}

func (r *MemoryRepository) UpdateSubmissionStatus(id int64, status model.SubmissionStatus, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	submission, ok := r.submissions[id]
	if !ok {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	submission.Status = status
	submission.ReviewNote = reviewNote
	submission.ReviewedBy = reviewedBy
	submission.UpdatedAt = time.Now()
	r.submissions[id] = submission
	return submission, nil
}

func (r *MemoryRepository) ApproveSubmission(id int64, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	submission, ok := r.submissions[id]
	if !ok {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	if submission.Status != model.SubmissionStatusPending {
		return model.PuzzleSubmission{}, ErrInvalidState
	}

	puzzle := submission.ToPuzzle()
	puzzle.ID = r.nextPuzzleID()
	r.puzzles[puzzle.ID] = puzzle
	r.rebuildPuzzlesBySet()

	submission.Status = model.SubmissionStatusApproved
	submission.ReviewNote = reviewNote
	submission.ReviewedBy = reviewedBy
	submission.UpdatedAt = time.Now()
	public := puzzle.Public()
	submission.ApprovedPuzzle = &public
	r.submissions[id] = submission
	return submission, nil
}

func (r *MemoryRepository) nextPuzzleID() int64 {
	var maxID int64
	for id := range r.puzzles {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func (r *MemoryRepository) nextSubmissionID() int64 {
	var maxID int64
	for id := range r.submissions {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

func (r *MemoryRepository) rebuildPuzzlesBySet() {
	r.puzzlesBySet = map[int64][]model.Puzzle{}
	for _, puzzle := range r.puzzles {
		r.puzzlesBySet[puzzle.PuzzleSetID] = append(r.puzzlesBySet[puzzle.PuzzleSetID], puzzle)
	}
	for setID, puzzles := range r.puzzlesBySet {
		sortPuzzles(puzzles)
		r.puzzlesBySet[setID] = puzzles
	}
}

func sortPuzzles(puzzles []model.Puzzle) {
	sort.Slice(puzzles, func(i, j int) bool {
		if puzzles[i].SortOrder == puzzles[j].SortOrder {
			return puzzles[i].ID < puzzles[j].ID
		}
		return puzzles[i].SortOrder < puzzles[j].SortOrder
	})
}

func sortSubmissions(submissions []model.PuzzleSubmission) {
	sort.Slice(submissions, func(i, j int) bool {
		return submissions[i].ID > submissions[j].ID
	})
}
