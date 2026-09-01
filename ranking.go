package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type rankingItem struct {
	UserName string
	Count    int
}

// ============================================================
// 今日排行
//
// mode:
// "all"     → 發言 + 貼圖
// "message" → 發言
// "sticker" → 貼圖
// ============================================================

func handleRanking(
	event *linebot.Event,
	mode string,
) {
	groupID := getGroupID(event)

	if groupID == "" {
		return
	}

	q := summaryQueue.ReadGroupInfo(groupID)

	// ========================================================
	// 今天 00:00
	// ========================================================

	now := time.Now().In(taipeiLocation)

	todayStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		taipeiLocation,
	)

	// ========================================================
	// 統計
	// ========================================================

	messageCounts := make(map[string]int)
	stickerCounts := make(map[string]int)

	for _, m := range q {

		messageTime := m.Time.In(taipeiLocation)

		// 只統計今天
		if messageTime.Before(todayStart) {
			continue
		}

		// ====================================================
		// 貼圖
		// ====================================================

		if m.MessageType == "sticker" {

			stickerCounts[m.UserName]++

			continue
		}

		// ====================================================
		// 發言
		//
		// MessageType == "" 是舊資料
		// 舊資料仍視為文字訊息
		// ====================================================

		if m.MessageType == "text" ||
			m.MessageType == "" {

			messageCounts[m.UserName]++
		}
	}

	// ========================================================
	// 建立回覆
	// ========================================================

	var reply strings.Builder

	switch mode {

	case "all":

		reply.WriteString("🐕 今天排行\n\n")

		reply.WriteString("💬 發言 TOP 10\n")
		reply.WriteString(formatRanking(
			messageCounts,
			"則",
		))

		reply.WriteString("\n🎨 貼圖 TOP 10\n")
		reply.WriteString(formatRanking(
			stickerCounts,
			"張",
		))

	case "message":

		reply.WriteString("💬 今日發言排行\n\n")

		reply.WriteString(formatRanking(
			messageCounts,
			"則",
		))

	case "sticker":

		reply.WriteString("🎨 今日貼圖排行\n\n")

		reply.WriteString(formatRanking(
			stickerCounts,
			"張",
		))
	}

	// ========================================================
	// 回覆 LINE
	// ========================================================

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply.String()),
	).Do(); err != nil {
		return
	}
}

// ============================================================
// 排序並產生排行榜文字
// ============================================================

func formatRanking(
	counts map[string]int,
	unit string,
) string {

	if len(counts) == 0 {
		return "目前還沒有資料～🐕\n"
	}

	items := make(
		[]rankingItem,
		0,
		len(counts),
	)

	for userName, count := range counts {

		items = append(
			items,
			rankingItem{
				UserName: userName,
				Count:    count,
			},
		)
	}

	// ========================================================
	// 排序
	//
	// 1. 數量由大到小
	// 2. 數量相同時按照名字排序
	// ========================================================

	sort.Slice(items, func(i, j int) bool {

		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}

		return items[i].UserName < items[j].UserName
	})

	// 只取 TOP 10

	if len(items) > 10 {
		items = items[:10]
	}

	// ========================================================
	// 產生排行榜文字
	// ========================================================

	var reply strings.Builder

	for i, item := range items {

		var rank string

		switch i {

		case 0:
			rank = "🥇"

		case 1:
			rank = "🥈"

		case 2:
			rank = "🥉"

		default:
			rank = fmt.Sprintf(
				"%d.",
				i+1,
			)
		}

		reply.WriteString(
			fmt.Sprintf(
				"%s %s　%d %s\n",
				rank,
				item.UserName,
				item.Count,
				unit,
			),
		)
	}

	return reply.String()
}

// ============================================================
// 取得使用者顯示名稱
//
// 群組：使用群組成員 Profile
// 聊天室：使用聊天室成員 Profile
// 一對一：使用一般 Profile
//
// 這樣可以避免群組中取得不到暱稱時
// 直接把 LINE User ID 存進排行榜。
// ============================================================

func getUserDisplayName(
	event *linebot.Event,
) string {

	userID := event.Source.UserID

	if userID == "" {
		return "未知使用者"
	}

	// ========================================================
	// 群組
	// ========================================================

	if event.Source.GroupID != "" {

		profile, err := bot.GetGroupMemberProfile(
			event.Source.GroupID,
			userID,
		).Do()

		if err == nil &&
			profile != nil &&
			profile.DisplayName != "" {

			return profile.DisplayName
		}

		return userID
	}

	// ========================================================
	// 聊天室
	// ========================================================

	if event.Source.RoomID != "" {

		profile, err := bot.GetRoomMemberProfile(
			event.Source.RoomID,
			userID,
		).Do()

		if err == nil &&
			profile != nil &&
			profile.DisplayName != "" {

			return profile.DisplayName
		}

		return userID
	}

	// ========================================================
	// 一對一
	// ========================================================

	profile, err := bot.GetProfile(
		userID,
	).Do()

	if err == nil &&
		profile != nil &&
		profile.DisplayName != "" {

		return profile.DisplayName
	}

	return userID
}

// ============================================================
// Store sticker message
// ============================================================

func handleStoreSticker(
	event *linebot.Event,
	message string,
) {

	userName := getUserDisplayName(event)

	m := MsgDetail{
		MsgText:     message,
		UserName:    userName,
		Time:        time.Now().In(taipeiLocation),
		MessageType: "sticker",
	}

	summaryQueue.AppendGroupInfo(
		getGroupID(event),
		m,
	)
}
