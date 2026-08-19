package main

import (
	"bufio"
	"math/rand"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func handleSiQuote(event *linebot.Event) {
	file, err := os.Open("斯語錄.txt")
	if err != nil {
		replyText(event, "斯語錄讀取失敗了嗷……")
		return
	}
	defer file.Close()

	var quotes []string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		quote := strings.TrimSpace(scanner.Text())

		// 空白行不算語錄
		if quote != "" {
			quotes = append(quotes, quote)
		}
	}

	if err := scanner.Err(); err != nil {
		replyText(event, "斯語錄讀取失敗了嗷……")
		return
	}

	if len(quotes) == 0 {
		replyText(event, "目前沒有斯語錄嗷……")
		return
	}

	quote := quotes[rand.Intn(len(quotes))]

	replyText(
		event,
		"🐕 斯語錄\n\n「"+quote+"」",
	)
}

func replyText(event *linebot.Event, text string) {
	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		// 忽略回覆錯誤
	}
}
