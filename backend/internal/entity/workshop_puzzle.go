package entity

// WorkshopPuzzle 是创意工坊查询所需的 puzzles 表只读投影。
// 该投影不执行 AutoMigrate，避免改变既有题库表结构。
type WorkshopPuzzle struct {
	ID          int64  `gorm:"column:id"`
	PuzzleSetID int64  `gorm:"column:puzzle_set_id"`
	AuthorName  string `gorm:"column:author_name"`
	HintImages  []byte `gorm:"column:hint_images"`
	Category    string `gorm:"column:category"`
	Difficulty  int    `gorm:"column:difficulty"`
	SortOrder   int    `gorm:"column:sort_order"`
}

func (WorkshopPuzzle) TableName() string { return "puzzles" }
