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

			// ========================================================
			// Draw
			// ========================================================

			if strings.Contains(text, ":draw") {

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
// 指令大全
// ========================================================

} else if strings.EqualFold(
	text,
	"指令大全",
) {

	handleCommandList(event)
				
			// ========================================================
			// 梗圖清單
			//
			// 只有完整輸入「梗圖清單」才觸發
			// ========================================================

			} else if strings.EqualFold(
				text,
				"梗圖清單",
			) && isGroupEvent(event) {

				handleMemeList(event)

			// ========================================================
			// 梗圖
			//
			// 梗圖
			// 梗圖 加班
			// 梗圖 傻眼
			//
			// 不會觸發：
			//
			// 梗圖Test
			// 梗圖機器人
			// 我想要梗圖
			// 這張梗圖很好笑
			// ========================================================

			} else if strings.EqualFold(
				text,
				"梗圖",
			) && isGroupEvent(event) {

				handleMeme(
					event,
					"",
				)

			} else if strings.HasPrefix(
				text,
				"梗圖 ",
			) && isGroupEvent(event) {

				keyword := strings.TrimSpace(
					strings.TrimPrefix(
						text,
						"梗圖 ",
					),
				)

				if keyword != "" {

					handleMeme(
						event,
						keyword,
					)
				}

			// ========================================================
			// 統整全部
			//
			// 統整全部
			// 統整全部 遊戲
			//
			// 一定要：
			//
			// 統整全部
			// 或
			// 統整全部 + 空格 + 自訂要求
			// ========================================================

			} else if strings.EqualFold(
				text,
				"統整全部",
			) && isGroupEvent(event) {

				handleSumAll(
					event,
					true,
					"",
				)

			} else if strings.HasPrefix(
				text,
				"統整全部 ",
			) && isGroupEvent(event) {

				customPrompt := strings.TrimSpace(
					strings.TrimPrefix(
						text,
						"統整全部 ",
					),
				)

				if customPrompt != "" {

					handleSumAll(
						event,
						true,
						customPrompt,
					)
				}

			// ========================================================
			// 統整
			//
			// 統整
			// 統整 遊戲
			// 統整 聚餐
			//
			// 不會觸發：
			//
			// 統整Test
			// 統整機器人
			// 統整一下
			// 我想統整
			// ========================================================

			} else if strings.EqualFold(
				text,
				"統整",
			) && isGroupEvent(event) {

				handleSumAll(
					event,
					false,
					"",
				)

			} else if strings.HasPrefix(
				text,
				"統整 ",
			) && isGroupEvent(event) {

				customPrompt := strings.TrimSpace(
					strings.TrimPrefix(
						text,
						"統整 ",
					),
				)

				if customPrompt != "" {

					handleSumAll(
						event,
						false,
						customPrompt,
					)
				}

			// ========================================================
			// 舊版統整指令
			//
			// 保留 :sum_all
			// ========================================================

			} else if strings.EqualFold(
				text,
				":sum_all",
			) && isGroupEvent(event) {

				handleSumAll(
					event,
					true,
					"",
				)

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

			// ========================================================
			// 貼圖兌換功能
			// ========================================================

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

			// ========================================================
			// 群組貼圖
			// ========================================================

			if isGroupEvent(event) {

				// 在群組中，只記錄貼圖，不回覆
				outStickerResult := fmt.Sprintf(
					"貼圖訊息: %s",
					kw,
				)

				handleStoreMsg(
					event,
					outStickerResult,
				)

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
	customPrompt string,
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

		startTime = summaryQueue.GetLastSummaryTime(
			groupID,
		)
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

	// 如果使用者沒有指定整理方向
	// 就使用一般摘要模式

	if strings.TrimSpace(customPrompt) == "" {

		customPrompt =
			"請依照聊天內容，自行判斷並整理最重要的討論、決定、日期時間地點，以及尚未確認的事情。"
	}

	// ========================================================
	// 第一段：固定 AI 整理規則
	// 第二段：使用者指定的整理方向
	// 第三段：實際聊天紀錄
	// ========================================================

	prompt := fmt.Sprintf(`
你是一隻很會整理 LINE 群組聊天紀錄的柴犬聊天機器人。

你的工作是幫群組成員快速了解這段聊天到底發生了什麼。

【固定整理規則】

請使用繁體中文。

你的目標不是把每一句話重新講一次，而是整理真正有價值的資訊。

請優先注意：

📌 大家正在討論什麼
💡 重要的討論內容
✅ 已經確定的事情
📅 日期、時間、地點
❓ 尚未決定或需要確認的事情
📝 有沒有明確的待辦事項

請不要逐一列出每個人的發言。

不要把聊天內容重新照順序重述。

如果只是閒聊，就簡單帶過。

不要為了湊內容而硬整理。

不要自行猜測聊天紀錄中沒有提到的事情。

不要把不確定的事情說成已經確定。

如果聊天紀錄沒有明確結論，就不要硬寫「最終結論」。

如果有明確結論，再自然地說明最後決定了什麼。

【使用者這次指定的整理方向】

%s

這是使用者針對這一次統整提出的額外要求。

請優先依照這個要求整理聊天內容。

如果使用者要求聚焦某個主題，就把與該主題相關的內容放在最前面。

如果使用者要求找出某些資訊，例如日期、時間、地點、人物意見或最後決定，就特別整理這些資訊。

但是不能因為使用者指定了方向，就自行補充聊天紀錄中沒有出現的資訊。

如果指定的主題在聊天紀錄中沒有相關內容，請直接說明沒有找到相關內容。

【輸出方式】

最重要的一點：

不要使用 Markdown 格式。

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

可以用 Emoji 分段，但不要讓格式太死板。

開頭可以使用：

「柴柴幫你整理好了，嗷～」

或：

「柴柴看了一下這段聊天～」

可以偶爾加入一點柴犬語氣，例如：

「嗷～」
「汪！」
「柴柴看了一下～」

但不要每一句都賣萌，也不要讓內容看起來幼稚。

整體內容請簡潔一點，讓群組成員可以快速看完。

請直接輸出整理結果。

不要解釋你使用了哪些規則。

【實際聊天紀錄】

%s
`, customPrompt, oriContext)

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
// GPT / Draw
//
// 目前只保留圖片生成功能。
// :gpt / :gpt4 已經移除。
// ============================================================

func handleGPT(
	action GPT_ACTIONS,
	event *linebot.Event,
	message string,
) {

	switch action {

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
// ============================================================
// Command List
// ============================================================

func handleCommandList(event *linebot.Event) {

	reply := `🐕 柴柴指令大全，嗷～

📖 聊天整理

統整
整理上一次統整後的新聊天內容。

統整 主題
針對指定主題整理。
例如：統整 遊戲

😂 梗圖

梗圖
隨機來一張梗圖。

梗圖 關鍵字
尋找指定類型的梗圖。
例如：梗圖 加班

梗圖清單
查看目前有哪些梗圖可以使用。


🎨 圖片

:draw
使用圖片生成功能。


🐕 其他

指令大全
查看目前所有可用指令。

柴柴目前還在努力學習更多功能，嗷～`

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply),
	).Do(); err != nil {

		log.Print(err)
	}
}
