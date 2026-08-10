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

			// GPT-4
			// 注意：要先判斷 :gpt4，
			// 否則 :gpt4 會先被 :gpt 判斷到。
			if strings.Contains(message.Text, ":gpt4") {

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

			// GPT
			} else if strings.Contains(message.Text, ":gpt") {

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
				"目前還沒有足夠的聊天紀錄可以整理喔，嗷～",
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
你是一隻很會整理聊天紀錄的柴犬聊天機器人。

請幫我整理下面這段 LINE 群組聊天紀錄。

你的目標不是把每一句話重新講一次，而是幫群組成員快速了解：
大家最近到底在聊什麼、討論出了什麼、有哪些事情已經決定，以及還有哪些事情需要確認。

聊天紀錄如下：

%s

請遵守以下規則：

請使用繁體中文。

最重要的一點：不要使用 Markdown 格式。

不要使用 # 標題。
不要使用 * 或 **。
不要使用 Markdown 條列符號。
不要使用 Markdown 表格。
不要使用 Markdown 程式碼區塊。

可以使用一般文字搭配 Emoji，例如：
📌
💡
✅
📅
❓
🐕
這些都可以。

請使用自然、口語、像朋友幫忙整理聊天內容的方式說話。

不要寫得像正式會議紀錄。

不要過度正式。

不要逐一列出每個人的發言。

請優先整理真正重要的內容。

如果大家有討論某個主題，請說明大家主要在討論什麼。

如果有明確決定的事情，請清楚說明。

如果有日期、時間、地點或活動，請整理出來。

如果有還沒有決定、需要再確認的事情，也請說明。

如果只是閒聊，就簡單帶過，不要為了湊內容而硬整理成重要事項。

不要自行猜測聊天紀錄中沒有提到的事情。

不要把不確定的事情說成已經確定。

整體內容請簡潔一點，讓群組成員可以快速看完。

開頭可以用類似：
「柴柴幫你整理好了，嗷～」
這種自然又有一點俏皮的方式。

可以偶爾加入一點柴犬的語氣，例如「嗷～」、「汪！」、「柴柴看了一下～」之類。

但不要每一句都賣萌，也不要讓內容看起來幼稚。

如果聊天紀錄沒有明確結論，就不要硬寫「最終結論」。

如果有明確結論，再自然地說明最後決定了什麼。

請直接輸出整理結果，不要解釋你使用了哪些規則。

`, oriContext)

	// ========================================================
	// Call Gemini
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
				linebot.NewTextMessage(
					"目前沒辦法幫你畫圖，嗷……",
				),
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
