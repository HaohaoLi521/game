package tests

import (
	"testing"

	"this-is-pun/backend/internal/data"
	"this-is-pun/backend/internal/model"
	"this-is-pun/backend/internal/service"
)

func newService() *service.GameService {
	return service.NewGameService(data.NewMemoryRepository(data.Seed()))
}

func attemptID(t *testing.T, game *service.GameService, puzzleID int64) string {
	t.Helper()
	puzzle, err := game.GetPublicPuzzle(puzzleID)
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.AttemptID == "" {
		t.Fatal("expected attempt id")
	}
	return puzzle.AttemptID
}

func TestCheckAnswerByText(t *testing.T) {
	game := newService()
	result, err := game.CheckAnswer(101, service.CheckAnswerRequest{
		AttemptID:  attemptID(t, game, 101),
		Answer:     "小鸟依人",
		AnswerMode: model.AnswerModeManual,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct || result.Score != 100 {
		t.Fatalf("expected correct with 100 score, got %+v", result)
	}
}

func TestCheckAnswerByAlias(t *testing.T) {
	game := newService()
	attempt := attemptID(t, game, 102)
	if _, err := game.GetHint(102, service.HintRequest{AttemptID: attempt, Level: 1}); err != nil {
		t.Fatal(err)
	}
	result, err := game.CheckAnswer(102, service.CheckAnswerRequest{
		AttemptID:  attempt,
		Answer:     "姜心比心",
		AnswerMode: model.AnswerModeTiles,
		HintsUsed:  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct || result.Score != 70 {
		t.Fatalf("expected alias correct with 70 score, got %+v", result)
	}
}

func TestCheckAnswerByPinyin(t *testing.T) {
	game := newService()
	result, err := game.CheckAnswer(110, service.CheckAnswerRequest{
		AttemptID:  attemptID(t, game, 110),
		Answer:     "羊眉吐气",
		AnswerMode: model.AnswerModeManual,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct {
		t.Fatalf("expected pinyin match, got %+v", result)
	}
}

func TestWrongAnswer(t *testing.T) {
	game := newService()
	result, err := game.CheckAnswer(101, service.CheckAnswerRequest{
		AttemptID: attemptID(t, game, 101),
		Answer:    "小鸟飞飞",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Correct || result.Answer != "" {
		t.Fatalf("expected wrong answer without leaked answer, got %+v", result)
	}
}

func TestCheckAnswerRequiresAttempt(t *testing.T) {
	game := newService()
	_, err := game.CheckAnswer(101, service.CheckAnswerRequest{
		Answer:     "小鸟依人",
		AnswerMode: model.AnswerModeManual,
	})
	if err == nil {
		t.Fatal("expected missing attempt to fail")
	}
}

func TestCheckAnswerRejectsUnsupportedMode(t *testing.T) {
	game := newService()
	_, err := game.CheckAnswer(101, service.CheckAnswerRequest{
		AttemptID:  attemptID(t, game, 101),
		Answer:     "小鸟依人",
		AnswerMode: "hacked",
	})
	if err == nil {
		t.Fatal("expected invalid answer mode to fail")
	}
}

func TestHintScoreComesFromServerAttempt(t *testing.T) {
	game := newService()
	attempt := attemptID(t, game, 101)
	if _, err := game.GetHint(101, service.HintRequest{AttemptID: attempt, Level: 2}); err != nil {
		t.Fatal(err)
	}
	result, err := game.CheckAnswer(101, service.CheckAnswerRequest{
		AttemptID:  attempt,
		Answer:     "小鸟依人",
		AnswerMode: model.AnswerModeManual,
		HintsUsed:  -99,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Correct || result.Score != 40 {
		t.Fatalf("expected server-side hint score 40, got %+v", result)
	}
}
