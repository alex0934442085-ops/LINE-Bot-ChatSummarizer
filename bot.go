package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, event := range events {
		if event.Type != linebot.EventTypeMessage {
			continue
		}

		switch message := event.Message.(type) {

		// =========================
		// Text Message
		// =========================
		case *linebot.TextMessage:

			// GPT
			if strings.Contains(message.Text, ":gpt") {

				if IsRedemptionEnabled() {
					if stickerRedeemable {
						handleGPT(GPT_Complete, event, message.Text)
						stickerRedeemable = false
					} else {
						handleRedeemRequestMsg(event)
					}
				} else {
					handleGPT(GPT_Complete, event, message.Text)
				}

			// GPT4
			} else if strings.Contains(message.Text, ":gpt4") {

				if IsRedemptionEnabled() {
					if stickerRedeemable {
						handleGPT(GPT_GPT4_Complete, event, message.Text)
						stickerRedeemable = false
					} else {
						handleRedeemRequestMsg(event)
					}
				} else {
					handleGPT(GPT_GPT4_Complete, event, message.Text)
				}

			// Draw
			} else if strings.Contains(message.Text, ":draw") {

				if IsRedemptionEnabled() {
					if stickerRedeemable {
						handleGPT(GPT_Draw, event, message.Text)
						stickerRedeemable = false
					} else {
						handleRedeemRequestMsg(event)
					}
				} else {
					handleGPT(GPT_Draw, event, message.Text)
				}

			// List group messages
			} else if strings.EqualFold(message.Text, ":list_all") && isGroupEvent(event) {

				handleListAll(event)

			// Summarize group messages
			} else if strings.EqualFold(message.Text, ":sum_all") && isGroupEvent(event) {

				handleSumAll(event)

			// Store group messages
			} else if isGroupEvent(event) {

				handleStoreMsg(event, message.Text)
			}

		// =========================
		// Sticker Message
		// =========================
		case *linebot.StickerMessage:

			var kw string

			for _, k := range message.Keywords {
				kw = kw + "," + k
			}

			log.Println(
				"Sticker: PID=",
				message.PackageID,
				"SID=",
				message.StickerID,
			)

			if IsRedemptionEnabled() {

				if message.PackageID == RedeemStickerPID &&
					message.StickerID == RedeemStickerSID {

					stickerRedeemable = true

					if _, err = bot.ReplyMessage(
						event.ReplyToken,
						linebot.NewTextMessage("你的賦能功能啟動了！"),
					).Do(); err != nil {
						log.Print(err)
					}
				}
			}

			if isGroupEvent(event) {

				// 在群組中，只記錄貼圖，不回覆
				outStickerResult := fmt.Sprintf(
					"貼圖訊息: %s",
					kw,
				)

				handleStoreMsg(event, outStickerResult)

			} else {

				// 一對一聊天則回覆
				outStickerResult := fmt.Sprintf(
					"貼圖訊息: %s, pkg: %s kw: %s text: %s",
					message.StickerID,
					message.PackageID,
					kw,
					message.Text,
				)

				if _, err = bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewTextMessage(outStickerResult),
				).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

// ============================================================
// Summarize all group messages
// ============================================================

func handleSumAll(event *linebot.Event) {

	// 取得目前群組的所有聊天紀錄
	oriContext := ""

	q := summaryQueue.ReadGroupInfo(getGroupID(event))

	for _, m := range q {

		oriContext += fmt.Sprintf(
			"[%s]: %s (%s)\n",
			m.UserName,
			m.MsgText,
			m.Time.Local().Format("2006-01-02 15:04:05"),
		)
	}

	// 如果目前沒有聊天紀錄
	if strings.TrimSpace(oriContext) == "" {

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"目前還沒有足夠的聊天紀錄可以整理喔。",
			),
		).Do(); err != nil {
			log.Print(err)
		}

		return
	}

	// ========================================================
	// AI Summary Prompt
	// ========================================================

	prompt := fmt.Sprintf(`
你是一個 LINE 群組聊天摘要機器人。

請將以下 LINE 群組聊天整理成簡潔、好讀的繁體中文摘要。

請優先整理：

1. 📌 討論主題
2. 💡 重要討論重點
3. ✅ 已經確定的事項
4. 📝 待辦事項
5. 📅 日期、時間、地點
6. ❓ 尚未決定或需要確認的事情

請不要逐句重述聊天內容。

如果聊天只是閒聊，請簡單說明，不要硬湊重點。

請使用條列式整理。

最後如果有明確結論，請加上：

「📢 最終結論：」

聊天紀錄：

%s
`, oriContext)

	// ========================================================
	// Call OpenAI
	// ========================================================

	reply := gptGPT3CompleteContext(prompt)

	// ========================================================
	// Reply directly to LINE group
	// ========================================================

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply),
	).Do(); err != nil {
		log.Print(err)
	}
}

// ============================================================
// List all group messages
// ============================================================

func handleListAll(event *linebot.Event) {

	reply := ""

	q := summaryQueue.ReadGroupInfo(getGroupID(event))

	for _, m := range q {

		reply += fmt.Sprintf(
			"[%s]: %s (%s)\n",
			m.UserName,
			m.MsgText,
			m.Time.Local().Format("2006-01-02 15:04:05"),
		)
	}

	if strings.TrimSpace(reply) == "" {
		reply = "目前沒有記錄到任何群組訊息。"
	}

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply),
	).Do(); err != nil {
		log.Print(err)
	}
}

// ============================================================
// GPT commands
// ============================================================

func handleGPT(
	action GPT_ACTIONS,
	event *linebot.Event,
	message string,
) {

	switch action {

	case GPT_Complete:

		reply := gptGPT3CompleteContext(message)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(reply),
		).Do(); err != nil {
			log.Print(err)
		}

	case GPT_GPT4_Complete:

		reply := gptGPT4CompleteContext(message)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(reply),
		).Do(); err != nil {
			log.Print(err)
		}

	case GPT_Draw:

		reply, err := gptImageCreate(message)

		if err != nil {

			if _, err := bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage("無法正確顯示圖形。"),
			).Do(); err != nil {
				log.Print(err)
			}

		} else {

			if _, err := bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage(
					"根據你的提示，畫出以下圖片：",
				),
				linebot.NewImageMessage(reply, reply),
			).Do(); err != nil {
				log.Print(err)
			}
		}
	}
}

// ============================================================
// Redemption
// ============================================================

func handleRedeemRequestMsg(event *linebot.Event) {

	userName := event.Source.UserID

	userProfile, err := bot.GetProfile(
		event.Source.UserID,
	).Do()

	if err == nil {
		userName = userProfile.DisplayName
	}

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(
			userName + ":你需要買貼圖，開啟這個功能",
		),
		linebot.NewStickerMessage(
			RedeemStickerPID,
			RedeemStickerSID,
		),
	).Do(); err != nil {
		log.Print(err)
	}
}

// ============================================================
// Store group message
// ============================================================

func handleStoreMsg(
	event *linebot.Event,
	message string,
) {

	userName := event.Source.UserID

	userProfile, err := bot.GetProfile(
		event.Source.UserID,
	).Do()

	if err == nil {
		userName = userProfile.DisplayName
	}

	m := MsgDetail{
		MsgText:  message,
		UserName: userName,
		Time:     time.Now(),
	}

	summaryQueue.AppendGroupInfo(
		getGroupID(event),
		m,
	)
}

// ============================================================
// Check whether event is from group / room
// ============================================================

func isGroupEvent(event *linebot.Event) bool {

	return event.Source.GroupID != "" ||
		event.Source.RoomID != ""
}

// ============================================================
// Get group ID
// ============================================================

func getGroupID(event *linebot.Event) string {

	if event.Source.GroupID != "" {
		return event.Source.GroupID
	}

	if event.Source.RoomID != "" {
		return event.Source.RoomID
	}

	return ""
}
