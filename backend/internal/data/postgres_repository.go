package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"this-is-pun/backend/internal/model"
)

const puzzleColumns = `
	id,
	puzzle_set_id,
	author_name,
	hint_images,
	hint_text,
	answer,
	answer_pinyin,
	answer_aliases,
	answer_length,
	candidate_chars,
	default_answer_mode,
	supported_answer_modes,
	blank_template,
	category,
	difficulty,
	explanation,
	sort_order`

const submissionColumns = `
	id,
	creator_name,
	contact,
	status,
	review_note,
	reviewed_by,
	puzzle_set_id,
	author_name,
	hint_images,
	hint_text,
	answer,
	answer_pinyin,
	answer_aliases,
	answer_length,
	candidate_chars,
	default_answer_mode,
	supported_answer_modes,
	blank_template,
	category,
	difficulty,
	explanation,
	sort_order,
	created_at,
	updated_at`

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(databaseURL string, seed SeedData) (*PostgresRepository, error) {
	if err := ensureDatabase(databaseURL); err != nil {
		return nil, err
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	repo := &PostgresRepository{db: db}
	if err := repo.ensureSchema(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := repo.seedIfEmpty(ctx, seed); err != nil {
		_ = db.Close()
		return nil, err
	}
	return repo, nil
}

func ensureDatabase(databaseURL string) error {
	targetURL, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}
	dbName := strings.TrimPrefix(targetURL.Path, "/")
	if dbName == "" {
		return errors.New("database name is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pingErr := db.PingContext(ctx)
		cancel()
		_ = db.Close()
		if pingErr == nil {
			return nil
		}
	}

	adminURL := *targetURL
	adminURL.Path = "/postgres"
	adminDB, err := sql.Open("pgx", adminURL.String())
	if err != nil {
		return err
	}
	defer adminDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := adminDB.PingContext(ctx); err != nil {
		return err
	}

	_, err = adminDB.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(dbName))
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return nil
}

func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

func (r *PostgresRepository) ensureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS puzzle_sets (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			domain_type TEXT NOT NULL DEFAULT '',
			cover_url TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS puzzles (
			id BIGSERIAL PRIMARY KEY,
			puzzle_set_id BIGINT NOT NULL REFERENCES puzzle_sets(id) ON DELETE CASCADE,
			author_name TEXT NOT NULL DEFAULT '',
			hint_images JSONB NOT NULL DEFAULT '[]'::jsonb,
			hint_text TEXT NOT NULL DEFAULT '',
			answer TEXT NOT NULL,
			answer_pinyin TEXT NOT NULL DEFAULT '',
			answer_aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
			answer_length INTEGER NOT NULL DEFAULT 0,
			candidate_chars JSONB NOT NULL DEFAULT '[]'::jsonb,
			default_answer_mode TEXT NOT NULL DEFAULT 'manual',
			supported_answer_modes JSONB NOT NULL DEFAULT '["manual", "tiles"]'::jsonb,
			blank_template TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			difficulty INTEGER NOT NULL DEFAULT 1,
			explanation TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			archived_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`ALTER TABLE puzzles ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ`,
		`CREATE INDEX IF NOT EXISTS idx_puzzles_set_sort ON puzzles (puzzle_set_id, sort_order, id)`,
		`CREATE INDEX IF NOT EXISTS idx_puzzles_active_set_sort ON puzzles (puzzle_set_id, sort_order, id) WHERE archived_at IS NULL`,
		`CREATE TABLE IF NOT EXISTS admin_users (
			id BIGSERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (
			token TEXT PRIMARY KEY,
			admin_user_id BIGINT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
			username TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires_at ON admin_sessions (expires_at)`,
		`CREATE TABLE IF NOT EXISTS puzzle_submissions (
			id BIGSERIAL PRIMARY KEY,
			creator_name TEXT NOT NULL DEFAULT '',
			contact TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'pending',
			review_note TEXT NOT NULL DEFAULT '',
			reviewed_by TEXT NOT NULL DEFAULT '',
			puzzle_set_id BIGINT NOT NULL DEFAULT 1,
			author_name TEXT NOT NULL DEFAULT '',
			hint_images JSONB NOT NULL DEFAULT '[]'::jsonb,
			hint_text TEXT NOT NULL DEFAULT '',
			answer TEXT NOT NULL,
			answer_pinyin TEXT NOT NULL DEFAULT '',
			answer_aliases JSONB NOT NULL DEFAULT '[]'::jsonb,
			answer_length INTEGER NOT NULL DEFAULT 0,
			candidate_chars JSONB NOT NULL DEFAULT '[]'::jsonb,
			default_answer_mode TEXT NOT NULL DEFAULT 'manual',
			supported_answer_modes JSONB NOT NULL DEFAULT '["manual", "tiles"]'::jsonb,
			blank_template TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			difficulty INTEGER NOT NULL DEFAULT 1,
			explanation TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_puzzle_submissions_status_id ON puzzle_submissions (status, id DESC)`,
	}
	for _, statement := range statements {
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) seedIfEmpty(ctx context.Context, seed SeedData) error {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM puzzle_sets`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, set := range seed.Sets {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO puzzle_sets (id, name, description, category, domain_type, cover_url)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			set.ID, set.Name, set.Description, set.Category, set.DomainType, set.CoverURL,
		); err != nil {
			return err
		}
	}
	for _, puzzle := range seed.Puzzles {
		if err := insertSeedPuzzle(ctx, tx, puzzle); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `SELECT setval(pg_get_serial_sequence('puzzle_sets', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM puzzle_sets), 1), true)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `SELECT setval(pg_get_serial_sequence('puzzles', 'id'), GREATEST((SELECT COALESCE(MAX(id), 1) FROM puzzles), 1), true)`); err != nil {
		return err
	}
	return tx.Commit()
}

func insertSeedPuzzle(ctx context.Context, tx *sql.Tx, puzzle model.Puzzle) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO puzzles (
			id,
			puzzle_set_id,
			author_name,
			hint_images,
			hint_text,
			answer,
			answer_pinyin,
			answer_aliases,
			answer_length,
			candidate_chars,
			default_answer_mode,
			supported_answer_modes,
			blank_template,
			category,
			difficulty,
			explanation,
			sort_order
		)
		VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8::jsonb, $9, $10::jsonb, $11, $12::jsonb, $13, $14, $15, $16, $17)`,
		puzzle.ID,
		puzzle.PuzzleSetID,
		puzzle.AuthorName,
		jsonParam(puzzle.HintImages),
		puzzle.HintText,
		puzzle.Answer,
		puzzle.AnswerPinyin,
		jsonParam(puzzle.AnswerAliases),
		puzzle.AnswerLength,
		jsonParam(puzzle.CandidateChars),
		string(puzzle.DefaultAnswerMode),
		jsonParam(puzzle.SupportedAnswerModes),
		puzzle.BlankTemplate,
		puzzle.Category,
		puzzle.Difficulty,
		puzzle.Explanation,
		puzzle.SortOrder,
	)
	return err
}

func (r *PostgresRepository) ListPuzzleSets() []model.PuzzleSet {
	rows, err := r.db.Query(`
		SELECT s.id, s.name, s.description, s.category, s.domain_type, s.cover_url, COUNT(p.id)::int AS puzzle_count
		FROM puzzle_sets s
		LEFT JOIN puzzles p ON p.puzzle_set_id = s.id AND p.archived_at IS NULL
		GROUP BY s.id
		ORDER BY s.id`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	sets := []model.PuzzleSet{}
	for rows.Next() {
		set, err := scanPuzzleSet(rows)
		if err == nil {
			sets = append(sets, set)
		}
	}
	return sets
}

func (r *PostgresRepository) GetPuzzleSet(id int64) (model.PuzzleSet, error) {
	row := r.db.QueryRow(`
		SELECT s.id, s.name, s.description, s.category, s.domain_type, s.cover_url, COUNT(p.id)::int AS puzzle_count
		FROM puzzle_sets s
		LEFT JOIN puzzles p ON p.puzzle_set_id = s.id AND p.archived_at IS NULL
		WHERE s.id = $1
		GROUP BY s.id`, id)
	set, err := scanPuzzleSet(row)
	if errors.Is(err, sql.ErrNoRows) {
		return model.PuzzleSet{}, ErrNotFound
	}
	return set, err
}

func (r *PostgresRepository) ListPuzzlesBySet(setID int64) ([]model.Puzzle, error) {
	if _, err := r.GetPuzzleSet(setID); err != nil {
		return nil, err
	}
	return r.queryPuzzles(`WHERE puzzle_set_id = $1 AND archived_at IS NULL ORDER BY sort_order, id`, setID)
}

func (r *PostgresRepository) GetPuzzle(id int64) (model.Puzzle, error) {
	puzzles, err := r.queryPuzzles(`WHERE id = $1 AND archived_at IS NULL`, id)
	if err != nil {
		return model.Puzzle{}, err
	}
	if len(puzzles) == 0 {
		return model.Puzzle{}, ErrNotFound
	}
	return puzzles[0], nil
}

func (r *PostgresRepository) ListAdminPuzzles() ([]model.Puzzle, error) {
	return r.queryPuzzles(`WHERE archived_at IS NULL ORDER BY puzzle_set_id, sort_order, id`)
}

func (r *PostgresRepository) CreatePuzzle(puzzle model.Puzzle) (model.Puzzle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	puzzle, err := insertPuzzle(ctx, r.db, puzzle)
	if err != nil {
		return model.Puzzle{}, err
	}
	return r.GetPuzzle(puzzle.ID)
}

type puzzleInserter interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func insertPuzzle(ctx context.Context, runner puzzleInserter, puzzle model.Puzzle) (model.Puzzle, error) {
	var id int64
	err := runner.QueryRowContext(ctx, `
		INSERT INTO puzzles (
			puzzle_set_id,
			author_name,
			hint_images,
			hint_text,
			answer,
			answer_pinyin,
			answer_aliases,
			answer_length,
			candidate_chars,
			default_answer_mode,
			supported_answer_modes,
			blank_template,
			category,
			difficulty,
			explanation,
			sort_order
		)
		VALUES ($1, $2, $3::jsonb, $4, $5, $6, $7::jsonb, $8, $9::jsonb, $10, $11::jsonb, $12, $13, $14, $15, $16)
		RETURNING id`,
		puzzle.PuzzleSetID,
		puzzle.AuthorName,
		jsonParam(puzzle.HintImages),
		puzzle.HintText,
		puzzle.Answer,
		puzzle.AnswerPinyin,
		jsonParam(puzzle.AnswerAliases),
		puzzle.AnswerLength,
		jsonParam(puzzle.CandidateChars),
		string(puzzle.DefaultAnswerMode),
		jsonParam(puzzle.SupportedAnswerModes),
		puzzle.BlankTemplate,
		puzzle.Category,
		puzzle.Difficulty,
		puzzle.Explanation,
		puzzle.SortOrder,
	).Scan(&id)
	if err != nil {
		return model.Puzzle{}, err
	}
	puzzle.ID = id
	return puzzle, nil
}

func (r *PostgresRepository) UpdatePuzzle(id int64, puzzle model.Puzzle) (model.Puzzle, error) {
	result, err := r.db.Exec(`
		UPDATE puzzles
		SET puzzle_set_id = $2,
			author_name = $3,
			hint_images = $4::jsonb,
			hint_text = $5,
			answer = $6,
			answer_pinyin = $7,
			answer_aliases = $8::jsonb,
			answer_length = $9,
			candidate_chars = $10::jsonb,
			default_answer_mode = $11,
			supported_answer_modes = $12::jsonb,
			blank_template = $13,
			category = $14,
			difficulty = $15,
			explanation = $16,
			sort_order = $17,
			updated_at = NOW()
		WHERE id = $1 AND archived_at IS NULL`,
		id,
		puzzle.PuzzleSetID,
		puzzle.AuthorName,
		jsonParam(puzzle.HintImages),
		puzzle.HintText,
		puzzle.Answer,
		puzzle.AnswerPinyin,
		jsonParam(puzzle.AnswerAliases),
		puzzle.AnswerLength,
		jsonParam(puzzle.CandidateChars),
		string(puzzle.DefaultAnswerMode),
		jsonParam(puzzle.SupportedAnswerModes),
		puzzle.BlankTemplate,
		puzzle.Category,
		puzzle.Difficulty,
		puzzle.Explanation,
		puzzle.SortOrder,
	)
	if err != nil {
		return model.Puzzle{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return model.Puzzle{}, err
	}
	if affected == 0 {
		return model.Puzzle{}, ErrNotFound
	}
	return r.GetPuzzle(id)
}

func (r *PostgresRepository) DeletePuzzle(id int64) error {
	result, err := r.db.Exec(`UPDATE puzzles SET archived_at = NOW(), updated_at = NOW() WHERE id = $1 AND archived_at IS NULL`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateAdminSession(session model.AdminSession) error {
	_, err := r.db.Exec(`
		INSERT INTO admin_sessions (token, admin_user_id, username, expires_at)
		VALUES ($1, $2, $3, $4)`,
		session.Token,
		session.UserID,
		session.Username,
		session.ExpiresAt,
	)
	return err
}

func (r *PostgresRepository) GetAdminSession(token string) (model.AdminSession, error) {
	var session model.AdminSession
	err := r.db.QueryRow(`
		SELECT token, admin_user_id, username, expires_at, created_at
		FROM admin_sessions
		WHERE token = $1`,
		token,
	).Scan(&session.Token, &session.UserID, &session.Username, &session.ExpiresAt, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AdminSession{}, ErrNotFound
	}
	return session, err
}

func (r *PostgresRepository) DeleteAdminSession(token string) error {
	_, err := r.db.Exec(`DELETE FROM admin_sessions WHERE token = $1`, token)
	return err
}

func (r *PostgresRepository) CreateAdminUser(user model.AdminUser) (model.AdminUser, error) {
	err := r.db.QueryRow(`
		INSERT INTO admin_users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, username, password_hash, created_at`,
		user.Username,
		user.PasswordHash,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return model.AdminUser{}, ErrAlreadyExists
		}
		return model.AdminUser{}, err
	}
	return user, nil
}

func (r *PostgresRepository) GetAdminUserByUsername(username string) (model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.QueryRow(`
		SELECT id, username, password_hash, created_at
		FROM admin_users
		WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.AdminUser{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) CreateSubmission(submission model.PuzzleSubmission) (model.PuzzleSubmission, error) {
	var id int64
	err := r.db.QueryRow(`
		INSERT INTO puzzle_submissions (
			creator_name,
			contact,
			status,
			review_note,
			reviewed_by,
			puzzle_set_id,
			author_name,
			hint_images,
			hint_text,
			answer,
			answer_pinyin,
			answer_aliases,
			answer_length,
			candidate_chars,
			default_answer_mode,
			supported_answer_modes,
			blank_template,
			category,
			difficulty,
			explanation,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12::jsonb, $13, $14::jsonb, $15, $16::jsonb, $17, $18, $19, $20, $21)
		RETURNING id`,
		submission.CreatorName,
		submission.Contact,
		string(model.SubmissionStatusPending),
		submission.ReviewNote,
		submission.ReviewedBy,
		submission.PuzzleSetID,
		submission.AuthorName,
		jsonParam(submission.HintImages),
		submission.HintText,
		submission.Answer,
		submission.AnswerPinyin,
		jsonParam(submission.AnswerAliases),
		submission.AnswerLength,
		jsonParam(submission.CandidateChars),
		string(submission.DefaultAnswerMode),
		jsonParam(submission.SupportedAnswerModes),
		submission.BlankTemplate,
		submission.Category,
		submission.Difficulty,
		submission.Explanation,
		submission.SortOrder,
	).Scan(&id)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	return r.GetSubmission(id)
}

func (r *PostgresRepository) ListSubmissions(status model.SubmissionStatus) ([]model.PuzzleSubmission, error) {
	if status == "" {
		return r.querySubmissions(`ORDER BY id DESC`)
	}
	return r.querySubmissions(`WHERE status = $1 ORDER BY id DESC`, string(status))
}

func (r *PostgresRepository) GetSubmission(id int64) (model.PuzzleSubmission, error) {
	submissions, err := r.querySubmissions(`WHERE id = $1`, id)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	if len(submissions) == 0 {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	return submissions[0], nil
}

func (r *PostgresRepository) UpdateSubmissionStatus(id int64, status model.SubmissionStatus, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error) {
	result, err := r.db.Exec(`
		UPDATE puzzle_submissions
		SET status = $2,
			review_note = $3,
			reviewed_by = $4,
			updated_at = NOW()
		WHERE id = $1`,
		id,
		string(status),
		reviewNote,
		reviewedBy,
	)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	if affected == 0 {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	return r.GetSubmission(id)
}

func (r *PostgresRepository) ApproveSubmission(id int64, reviewNote string, reviewedBy string) (model.PuzzleSubmission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	defer tx.Rollback()

	submission, err := scanSubmission(tx.QueryRowContext(ctx, `SELECT `+submissionColumns+` FROM puzzle_submissions WHERE id = $1 FOR UPDATE`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.PuzzleSubmission{}, ErrNotFound
	}
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	if submission.Status != model.SubmissionStatusPending {
		return model.PuzzleSubmission{}, ErrInvalidState
	}

	puzzle, err := insertPuzzle(ctx, tx, submission.ToPuzzle())
	if err != nil {
		return model.PuzzleSubmission{}, err
	}

	submission, err = scanSubmission(tx.QueryRowContext(ctx, `
		UPDATE puzzle_submissions
		SET status = $2,
			review_note = $3,
			reviewed_by = $4,
			updated_at = NOW()
		WHERE id = $1
		RETURNING `+submissionColumns,
		id,
		string(model.SubmissionStatusApproved),
		reviewNote,
		reviewedBy,
	))
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.PuzzleSubmission{}, err
	}

	public := puzzle.Public()
	submission.ApprovedPuzzle = &public
	return submission, nil
}

func (r *PostgresRepository) queryPuzzles(tail string, args ...any) ([]model.Puzzle, error) {
	rows, err := r.db.Query(`SELECT `+puzzleColumns+` FROM puzzles `+tail, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	puzzles := []model.Puzzle{}
	for rows.Next() {
		puzzle, err := scanPuzzle(rows)
		if err != nil {
			return nil, err
		}
		puzzles = append(puzzles, puzzle)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(puzzles, func(i, j int) bool {
		if puzzles[i].PuzzleSetID == puzzles[j].PuzzleSetID {
			if puzzles[i].SortOrder == puzzles[j].SortOrder {
				return puzzles[i].ID < puzzles[j].ID
			}
			return puzzles[i].SortOrder < puzzles[j].SortOrder
		}
		return puzzles[i].PuzzleSetID < puzzles[j].PuzzleSetID
	})
	return puzzles, nil
}

func (r *PostgresRepository) querySubmissions(tail string, args ...any) ([]model.PuzzleSubmission, error) {
	rows, err := r.db.Query(`SELECT `+submissionColumns+` FROM puzzle_submissions `+tail, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := []model.PuzzleSubmission{}
	for rows.Next() {
		submission, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPuzzleSet(row scanner) (model.PuzzleSet, error) {
	var set model.PuzzleSet
	err := row.Scan(
		&set.ID,
		&set.Name,
		&set.Description,
		&set.Category,
		&set.DomainType,
		&set.CoverURL,
		&set.PuzzleCount,
	)
	return set, err
}

func scanPuzzle(row scanner) (model.Puzzle, error) {
	var puzzle model.Puzzle
	var hintImages []byte
	var aliases []byte
	var candidates []byte
	var supportedModes []byte
	var defaultMode string
	err := row.Scan(
		&puzzle.ID,
		&puzzle.PuzzleSetID,
		&puzzle.AuthorName,
		&hintImages,
		&puzzle.HintText,
		&puzzle.Answer,
		&puzzle.AnswerPinyin,
		&aliases,
		&puzzle.AnswerLength,
		&candidates,
		&defaultMode,
		&supportedModes,
		&puzzle.BlankTemplate,
		&puzzle.Category,
		&puzzle.Difficulty,
		&puzzle.Explanation,
		&puzzle.SortOrder,
	)
	if err != nil {
		return model.Puzzle{}, err
	}
	puzzle.DefaultAnswerMode = model.AnswerMode(defaultMode)
	if err := decodeJSON(hintImages, &puzzle.HintImages); err != nil {
		return model.Puzzle{}, err
	}
	if err := decodeJSON(aliases, &puzzle.AnswerAliases); err != nil {
		return model.Puzzle{}, err
	}
	if err := decodeJSON(candidates, &puzzle.CandidateChars); err != nil {
		return model.Puzzle{}, err
	}
	if err := decodeJSON(supportedModes, &puzzle.SupportedAnswerModes); err != nil {
		return model.Puzzle{}, err
	}
	return puzzle, nil
}

func scanSubmission(row scanner) (model.PuzzleSubmission, error) {
	var submission model.PuzzleSubmission
	var status string
	var hintImages []byte
	var aliases []byte
	var candidates []byte
	var supportedModes []byte
	var defaultMode string
	err := row.Scan(
		&submission.ID,
		&submission.CreatorName,
		&submission.Contact,
		&status,
		&submission.ReviewNote,
		&submission.ReviewedBy,
		&submission.PuzzleSetID,
		&submission.AuthorName,
		&hintImages,
		&submission.HintText,
		&submission.Answer,
		&submission.AnswerPinyin,
		&aliases,
		&submission.AnswerLength,
		&candidates,
		&defaultMode,
		&supportedModes,
		&submission.BlankTemplate,
		&submission.Category,
		&submission.Difficulty,
		&submission.Explanation,
		&submission.SortOrder,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return model.PuzzleSubmission{}, err
	}
	submission.Status = model.SubmissionStatus(status)
	submission.DefaultAnswerMode = model.AnswerMode(defaultMode)
	if err := decodeJSON(hintImages, &submission.HintImages); err != nil {
		return model.PuzzleSubmission{}, err
	}
	if err := decodeJSON(aliases, &submission.AnswerAliases); err != nil {
		return model.PuzzleSubmission{}, err
	}
	if err := decodeJSON(candidates, &submission.CandidateChars); err != nil {
		return model.PuzzleSubmission{}, err
	}
	if err := decodeJSON(supportedModes, &submission.SupportedAnswerModes); err != nil {
		return model.PuzzleSubmission{}, err
	}
	return submission, nil
}

func jsonParam(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

func decodeJSON(bytes []byte, target any) error {
	if len(bytes) == 0 || string(bytes) == "null" {
		bytes = []byte("[]")
	}
	return json.Unmarshal(bytes, target)
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func FormatDatabaseURL(host string, port int, dbName string, user string, password string) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", url.QueryEscape(user), url.QueryEscape(password), host, port, dbName)
}
