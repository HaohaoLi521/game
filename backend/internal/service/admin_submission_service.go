package service

import (
	"fmt"
	"strings"

	"this-is-pun/backend/internal/model"
)

// BatchReviewAction 定义管理员批量审核动作。
type BatchReviewAction string

const (
	// BatchReviewApprove 批量通过投稿并入库。
	BatchReviewApprove BatchReviewAction = "approve"
	// BatchReviewReject 批量拒绝投稿。
	BatchReviewReject BatchReviewAction = "reject"
)

// BatchReviewInput 是批量审核用例的输入 DTO。
type BatchReviewInput struct {
	SubmissionIDs []int64
	Action        BatchReviewAction
	ReviewNote    string
	Reviewer      string
}

// AdminSubmissionService 聚合管理端投稿审核用例。
type AdminSubmissionService struct{ game *GameService }

func NewAdminSubmissionService(game *GameService) *AdminSubmissionService {
	return &AdminSubmissionService{game: game}
}

// BatchReview 先校验整批投稿均可审核，再逐条执行既有审核业务。
func (s *AdminSubmissionService) BatchReview(input BatchReviewInput) ([]model.PuzzleSubmission, error) {
	if len(input.SubmissionIDs) == 0 || (input.Action != BatchReviewApprove && input.Action != BatchReviewReject) {
		return nil, ErrInvalidRequest
	}
	seen := make(map[int64]struct{}, len(input.SubmissionIDs))
	for _, id := range input.SubmissionIDs {
		if id < 1 {
			return nil, ErrInvalidRequest
		}
		if _, exists := seen[id]; exists {
			return nil, ErrInvalidRequest
		}
		seen[id] = struct{}{}
		submission, err := s.game.repo.GetSubmission(id)
		if err != nil {
			return nil, convertRepoError(err)
		}
		if submission.Status != model.SubmissionStatusPending {
			return nil, ErrInvalidRequest
		}
	}

	items := make([]model.PuzzleSubmission, 0, len(input.SubmissionIDs))
	req := SubmissionReviewRequest{ReviewNote: strings.TrimSpace(input.ReviewNote)}
	for _, id := range input.SubmissionIDs {
		var (
			submission model.PuzzleSubmission
			err        error
		)
		if input.Action == BatchReviewApprove {
			submission, err = s.game.ApproveSubmission(id, input.Reviewer, req)
		} else {
			submission, err = s.game.RejectSubmission(id, input.Reviewer, req)
		}
		if err != nil {
			return nil, fmt.Errorf("batch review submission %d: %w", id, err)
		}
		items = append(items, submission)
	}
	return items, nil
}
