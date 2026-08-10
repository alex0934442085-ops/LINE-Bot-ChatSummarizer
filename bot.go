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

		// ============================================================
		// Text Message
		// ============================================================

		case *linebot.TextMessage:

			text := strings.TrimSpace(message.Text)

			// GPT-4
			if strings.Contains(text, ":gpt4") {

				if IsRedemptionEnabled() {

					if stickerRedeemable {

						handleGPT(
							GPT_GPT4_Complete,
							event,
							text,
						)

						stickerRedeemable = false

					} else {

						handleRedeemRequestMsg(event)
					}

				} else {

					handleGPT(
						GPT_GPT4_Complete,
						event,
						text,
					)
				}

			// GPT
			} else if strings.Contains(text, ":gpt") {

				if IsRedemptionEnabled() {

					if stickerRedeemable {

						handleGPT(
							GPT_Complete,
							event,
							text,
						)

						stickerRedeemable = false

					} else {

						handleRedeemRequestMsg(event)
					}

				} else {

					handleGPT(
						GPT_Complete,
						event,
						text,
					)
				}

			// Draw
			} else if strings.Contains(text, ":draw") {

				if IsRedemptionEnabled() {

					if stickerRedeemable {

						handleGPT(
							GPT_Draw,
							event,
							text,
						)

						stickerRedeemable = false

					} else {

						handleRedeemRequestMsg(event)
					}

				} else {

					handleGPT(
						GPT_Draw,
						event,
						text,
					)
				}

			// ========================================================
			// 統整
			// ========================================================

			} else if strings.EqualFold(text, "統整") &&
				isGroupEvent(event) {

				handleSumAll(event, false)

			// ========================================================
			// 統整全部
			// ========================================================

			} else if strings.EqualFold(text, "統整全部") &&
				isGroupEvent(event) {

				handleSumAll(event, true)

			// ========================================================
			// 舊指令 :sum_all
			// 保留給你使用
			// ========================================================

			} else if strings.EqualFold(text, ":sum_all") &&
				isGroupEvent(event) {

				handleSumAll(event, true)

			// ========================================================
			// 顯示全部聊天紀錄
			// ========================================================

			} else if strings.EqualFold(text, ":list_all") &&
				isGroupEvent(event) {

				handleListAll(event)

			// ========================================================
			// Store group messages
			// ========================================================

			} else if isGroupEvent(event) {

				handleStoreMsg(
					event,
					text,
				)
			}

		// ============================================================
		// Sticker Message
		// ============================================================

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
						linebot.NewTextMessage(
							"你的賦能功能啟動了！",
						),
					).Do(); err != nil {

						log.Print(err)
					}
				}
			}

			if isGroupEvent(event) {

				outStickerResult := fmt.Sprintf(
					"貼圖訊息: %s",
					kw,
				)

				handleStoreMsg(
					event,
					outStickerResult,
				)

			} else {

				outStickerResult := fmt.Sprintf(
					"貼圖訊息: %s, pkg: %s kw: %s text: %s",
					message.StickerID,
					message.PackageID,
					kw,
					message.Text,
				)

				if _, err = bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewTextMessage(
						outStickerResult,
					),
				).Do(); err != nil {

					log.Print(err)
				}
			}
		}
	}
}

// ============================================================
// Summarize group messages
//
// full = false
// → 只整理上一次統整之後的訊息
//
// full = true
// → 整理全部訊息
// ============================================================

func handleSumAll(
	event *linebot.Event,
	full bool,
) {

	groupID := getGroupID(event)

	if groupID == "" {
		return
	}

	// ========================================================
	// 取得全部聊天紀錄
	// ========================================================

	q := summaryQueue.ReadGroupInfo(groupID)

	if len(q) == 0 {

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"目前還沒有聊天紀錄可以整理喔，嗷～",
			),
		).Do(); err != nil {

			log.Print(err)
		}

		return
	}

	// ========================================================
	// 決定本次要整理的起點
	// ========================================================

	var startTime time.Time

	if !full {

		startTime = summaryQueue.GetLastSummaryTime(groupID)
	}

	// ========================================================
	// 建立本次摘要用的聊天內容
	// ========================================================

	oriContext := ""

	var latestMessageTime time.Time

	for _, m := range q {

		// 統整模式：
		// 只抓上次統整之後的訊息
		if !full &&
			!startTime.IsZero() &&
			!m.Time.After(startTime) {

			continue
		}

		oriContext += fmt.Sprintf(
			"[%s]: %s (%s)\n",
			m.UserName,
			m.MsgText,
			m.Time.Local().Format(
				"2006-01-02 15:04:05",
			),
		)

		if m.Time.After(latestMessageTime) {
			latestMessageTime = m.Time
		}
	}

	// ========================================================
	// 沒有新訊息
	// ========================================================

	if strings.TrimSpace(oriContext) == "" {

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"柴柴看了一下～上次統整之後好像沒有新的聊天內容耶，嗷？🐕",
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
你是一隻很會整理 LINE 群組聊天紀錄的柴犬聊天機器人。

請幫我整理下面這一段 LINE 群組聊天。

你的目標不是把每一句話重新講一次，而是讓群組成員快速知道：
大家在聊什麼、重要的事情是什麼、做了哪些決定，以及還有什麼事情需要確認。

聊天紀錄：

%s

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

請使用自然、口語、像朋友幫忙整理聊天內容的方式說話。

不要寫得像正式會議紀錄。
不要過度正式。
不要逐一列出每個人的發言。

請優先整理真正重要的內容。

如果大家有討論某個主題，請說明主要在討論什麼。

如果有明確決定的事情，請清楚說明。

如果有日期、時間、地點或活動，請整理出來。

如果有還沒有決定、需要再確認的事情，也請說明。

如果只是閒聊，就簡單帶過，不要為了湊內容而硬整理。

不要自行猜測聊天紀錄中沒有提到的事情。

不要把不確定的事情說成已經確定。

整體內容請簡潔一點，讓群組成員可以快速看完。

開頭可以用類似：
「柴柴幫你整理好了，嗷～」
這種自然又有一點俏皮的方式。

可以偶爾加入一點柴犬的語氣，例如：
「嗷～」
「汪！」
「柴柴看了一下～」

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
	// Gemini 發生錯誤
	// 不要更新統整時間
	// ========================================================

	if strings.HasPrefix(reply, "Err:") {

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(reply),
		).Do(); err != nil {

			log.Print(err)
		}

		return
	}

	// ========================================================
	// 回覆群組
	// ========================================================

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply),
	).Do(); err != nil {

		log.Print(err)
		return
	}

	// ========================================================
	// AI 成功 + LINE 成功
	// 才記錄這次統整的位置
	// ========================================================

	if !latestMessageTime.IsZero() {

		summaryQueue.SetLastSummaryTime(
			groupID,
			latestMessageTime,
		)

		log.Println(
			"Summary completed. Group:",
			groupID,
			"Time:",
			latestMessageTime,
		)
	}
}

// ============================================================
// List all group messages
// ============================================================

func handleListAll(event *linebot.Event) {

	reply := ""

	q := summaryQueue.ReadGroupInfo(
		getGroupID(event),
	)

	for _, m := range q {

		reply += fmt.Sprintf(
			"[%s]: %s (%s)\n",
			m.UserName,
			m.MsgText,
			m.Time.Local().Format(
				"2006-01-02 15:04:05",
			),
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

		reply := gptGPT3CompleteContext(
			message,
		)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(reply),
		).Do(); err != nil {

			log.Print(err)
		}

	case GPT_GPT4_Complete:

		reply := gptGPT4CompleteContext(
			message,
		)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(reply),
		).Do(); err != nil {

			log.Print(err)
		}

	case GPT_Draw:

		reply, err := gptImageCreate(
			message,
		)

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
				linebot.NewImageMessage(
					reply,
					reply,
				),
			).Do(); err != nil {

				log.Print(err)
			}
		}
	}
}

// ============================================================
// Redemption
// ============================================================

func handleRedeemRequestMsg(
	event *linebot.Event,
) {

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
			userName +
				":你需要買貼圖，開啟這個功能",
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

func isGroupEvent(
	event *linebot.Event,
) bool {

	return event.Source.GroupID != "" ||
		event.Source.RoomID != ""
}

// ============================================================
// Get group ID
// ============================================================

func getGroupID(
	event *linebot.Event,
) string {

	if event.Source.GroupID != "" {
		return event.Source.GroupID
	}

	if event.Source.RoomID != "" {
		return event.Source.RoomID
	}

	return ""
}
