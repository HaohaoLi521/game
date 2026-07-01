package data

import (
	"fmt"
	"unicode/utf8"

	"github.com/mozillazg/go-pinyin"

	"this-is-pun/backend/internal/model"
)

func Seed() SeedData {
	puzzles := []model.Puzzle{
		puzzle(101, "小鸟依人", "xiao niao yi ren", nil, "成语", "小鸟 + 依人，组成成语小鸟依人。", "emoji:🐦", "这是小鸟", "emoji:🧍", "这是____", "小 鸟 依 人 心 比 大 江 可 爱", model.AnswerModeTiles),
		puzzle(102, "将心比心", "jiang xin bi xin", []string{"姜心比心"}, "成语", "姜和将同音，姜在比心就是将心比心。", "emoji:🫚", "这是姜", "emoji:🫰", "姜在比心", "将 心 比 心 姜 情 星 花 想 明", model.AnswerModeTiles),
		puzzle(103, "一见钟情", "yi jian zhong qing", []string{"一箭钟情"}, "成语", "一支箭遇见钟和爱心，谐音一见钟情。", "emoji:🏹", "这是一箭", "emoji:🔔", "这是钟情", "一 见 钟 情 箭 中 晴 心 爱 上", model.AnswerModeManual),
		puzzle(104, "花好月圆", "hua hao yue yuan", nil, "成语", "花很好，月亮很圆，合起来是花好月圆。", "emoji:🌸", "花很好", "emoji:🌕", "月亮很圆", "花 好 月 圆 草 亮 满 星 香 甜", model.AnswerModeTiles),
		puzzle(105, "骑虎难下", "qi hu nan xia", nil, "成语", "骑在老虎上，不好下来。", "emoji:🐯", "这是老虎", "emoji:🏇", "骑上去了", "骑 虎 难 下 马 上 来 去 跑 跳", model.AnswerModeManual),
		puzzle(106, "画蛇添足", "hua she tian zu", nil, "成语", "蛇被画出来之后又添了脚。", "emoji:🐍", "这是蛇", "emoji:🦶", "又添了足", "画 蛇 添 足 花 舌 天 走 手 笔", model.AnswerModeTiles),
		puzzle(107, "守株待兔", "shou zhu dai tu", nil, "成语", "守着树桩等待兔子。", "emoji:🪵", "这是树桩", "emoji:🐰", "等兔子来", "守 株 待 兔 手 住 带 图 木 林", model.AnswerModeTiles),
		puzzle(108, "杯弓蛇影", "bei gong she ying", nil, "成语", "杯子里出现弓和蛇的影子。", "emoji:🥤", "这是杯子", "emoji:🏹", "杯里有弓影", "杯 弓 蛇 影 背 工 设 景 水 月", model.AnswerModeManual),
		puzzle(109, "鸡飞狗跳", "ji fei gou tiao", nil, "成语", "鸡飞起来，狗跳起来。", "emoji:🐔", "鸡飞了", "emoji:🐶", "狗跳了", "鸡 飞 狗 跳 机 非 够 条 鸟 跑", model.AnswerModeTiles),
		puzzle(110, "羊眉吐气", "yang mei tu qi", []string{"扬眉吐气"}, "成语", "羊的眉毛 + 吐气，谐音扬眉吐气。", "emoji:🐑", "羊的眉毛", "emoji:💨", "吐气", "扬 眉 吐 气 羊 美 土 七 风 口", model.AnswerModeManual),
		puzzle(111, "藕断丝连", "ou duan si lian", nil, "成语", "莲藕断开后仍有丝相连。", "emoji:🪷", "这是藕", "emoji:🧵", "丝还连着", "藕 断 丝 连 偶 段 思 莲 线 心", model.AnswerModeTiles),
		puzzle(112, "马到成功", "ma dao cheng gong", nil, "成语", "马到了，成功也来了。", "emoji:🐴", "马到了", "emoji:🏆", "成功啦", "马 到 成 功 吗 倒 城 工 快 赢", model.AnswerModeManual),
		puzzle(113, "三心二意", "san xin er yi", nil, "成语", "三颗心加二个想法。", "emoji:💗", "三颗心", "emoji:✌️", "二个意思", "三 心 二 意 山 新 耳 一 想 爱", model.AnswerModeTiles),
		puzzle(114, "十全十美", "shi quan shi mei", nil, "成语", "十个都全，十个都美。", "emoji:🔟", "这是十", "emoji:💯", "很完美", "十 全 十 美 石 泉 是 每 好 满", model.AnswerModeManual),
		puzzle(115, "眉开眼笑", "mei kai yan xiao", nil, "成语", "眉毛舒展，眼睛笑起来。", "emoji:👁️", "眼睛在笑", "emoji:😊", "笑起来", "眉 开 眼 笑 美 凯 言 小 脸 乐", model.AnswerModeTiles),
		puzzle(116, "井井有条", "jing jing you tiao", nil, "成语", "井一口又一口，旁边有条纹。", "emoji:#️⃣", "像一口井", "emoji:〰️", "有条纹", "井 井 有 条 景 经 油 跳 线 路", model.AnswerModeManual),
		puzzle(117, "口是心非", "kou shi xin fei", nil, "成语", "嘴上说是，心里却不是。", "emoji:👄", "口说是", "emoji:💔", "心里非", "口 是 心 非 扣 事 新 飞 嘴 爱", model.AnswerModeTiles),
		puzzle(118, "雪中送炭", "xue zhong song tan", nil, "成语", "雪天里送来炭火。", "emoji:❄️", "雪中", "emoji:🪨", "送来炭", "雪 中 送 炭 学 钟 松 坛 火 冬", model.AnswerModeManual),
		puzzle(119, "鱼跃龙门", "yu yue long men", nil, "成语", "鱼跃起来，跳过龙门。", "emoji:🐟", "鱼跃起", "emoji:🐉", "龙门", "鱼 跃 龙 门 雨 月 隆 们 水 高", model.AnswerModeTiles),
		puzzle(120, "心花怒放", "xin hua nu fang", nil, "成语", "心里的花开心地盛放。", "emoji:💖", "心里有花", "emoji:🌼", "怒放", "心 花 怒 放 新 华 女 方 开 香", model.AnswerModeManual),
	}

	return SeedData{
		Sets: []model.PuzzleSet{
			{
				ID:          1,
				Name:        "主线热身题库",
				Description: "20 道适合验证双模式玩法的谐音梗题目。",
				Category:    "main",
				DomainType:  "成语",
				CoverURL:    "emoji:👟",
				PuzzleCount: len(puzzles),
			},
		},
		Puzzles: puzzles,
	}
}

func puzzle(id int64, answer string, pinyin string, aliases []string, category string, explanation string, imageOne string, labelOne string, imageTwo string, labelTwo string, candidates string, mode model.AnswerMode) model.Puzzle {
	runes := []rune(candidates)
	chars := make([]model.CandidateChar, 0, len(runes))
	index := 1
	for _, char := range runes {
		if char == ' ' {
			continue
		}
		chars = append(chars, model.CandidateChar{
			ID:     fmt.Sprintf("p%d-c%d", id, index),
			Char:   string(char),
			Pinyin: charPinyin(string(char)),
		})
		index++
	}

	return model.Puzzle{
		ID:          id,
		PuzzleSetID: 1,
		AuthorName:  "QQ",
		HintImages: []model.HintImage{
			{ID: fmt.Sprintf("p%d-a", id), URL: imageOne, Label: labelOne, Alt: labelOne},
			{ID: fmt.Sprintf("p%d-b", id), URL: imageTwo, Label: labelTwo, Alt: labelTwo},
		},
		Answer:               answer,
		AnswerPinyin:         pinyin,
		AnswerAliases:        aliases,
		AnswerLength:         utf8.RuneCountInString(answer),
		CandidateChars:       chars,
		DefaultAnswerMode:    mode,
		SupportedAnswerModes: []model.AnswerMode{model.AnswerModeManual, model.AnswerModeTiles},
		BlankTemplate:        "这是" + underline(utf8.RuneCountInString(answer)),
		Category:             category,
		Difficulty:           1,
		Explanation:          explanation,
		SortOrder:            int(id - 100),
	}
}

func underline(length int) string {
	out := ""
	for i := 0; i < length; i++ {
		out += "_"
	}
	return out
}

func charPinyin(input string) string {
	args := pinyin.NewArgs()
	args.Style = pinyin.Normal
	parts := pinyin.Pinyin(input, args)
	if len(parts) == 0 || len(parts[0]) == 0 {
		return ""
	}
	return parts[0][0]
}
