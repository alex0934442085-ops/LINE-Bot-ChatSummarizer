package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// ============================================================
// Meme 功能
//
// 梗圖
// → 隨機抽一張梗圖
//
// 梗圖 加班
// → 從檔名包含「加班」的梗圖中隨機抽一張
//
// 完全不使用 AI，不消耗 Gemini Token。
// ============================================================

func handleMeme(event *linebot.Event, keyword string) {

	keyword = strings.TrimSpace(keyword)

	// ========================================================
	// 讀取 meme 資料夾
	// ========================================================

	memeDir := "meme"

	files, err := os.ReadDir(memeDir)

	if err != nil {

		log.Println("Read meme directory error:", err)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"柴柴找不到梗圖資料夾，嗷……🐕",
			),
		).Do(); err != nil {
			log.Print(err)
		}

		return
	}

	// ========================================================
	// 找符合條件的圖片
	// ========================================================

	var images []string

	for _, file := range files {

		if file.IsDir() {
			continue
		}

		filename := file.Name()

		ext := strings.ToLower(
			filepath.Ext(filename),
		)

		// 只接受圖片
		if ext != ".jpg" &&
			ext != ".jpeg" &&
			ext != ".png" &&
			ext != ".webp" {

			continue
		}

		// 有指定關鍵字時，只找檔名包含關鍵字的圖片
		if keyword != "" &&
			!strings.Contains(
				strings.ToLower(filename),
				strings.ToLower(keyword),
			) {

			continue
		}

		images = append(
			images,
			filename,
		)
	}

	// ========================================================
	// 沒找到圖片
	// ========================================================

	if len(images) == 0 {

		if keyword == "" {

			if _, err := bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage(
					"目前還沒有梗圖可以丟給你，嗷……🐕",
				),
			).Do(); err != nil {
				log.Print(err)
			}

		} else {

			if _, err := bot.ReplyMessage(
				event.ReplyToken,
				linebot.NewTextMessage(
					fmt.Sprintf(
						"柴柴找不到「%s」相關的梗圖耶……🐕",
						keyword,
					),
				),
			).Do(); err != nil {
				log.Print(err)
			}
		}

		return
	}

	// ========================================================
	// 隨機選一張
	// ========================================================

	rand.Seed(time.Now().UnixNano())

	selected := images[
		rand.Intn(len(images)),
	]

	// ========================================================
	// 建立 GitHub Raw 圖片網址
	// ========================================================

	imageURL := buildMemeURL(selected)

	log.Println(
		"Meme selected:",
		selected,
	)

	// ========================================================
	// 回覆 LINE
	// ========================================================

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewImageMessage(
			imageURL,
			imageURL,
		),
	).Do(); err != nil {

		log.Println(
			"Meme reply error:",
			err,
		)
	}
}

// ============================================================
// 建立 GitHub Raw URL
// ============================================================

func buildMemeURL(filename string) string {

	return "https://raw.githubusercontent.com/alex0934442085-ops/LINE-Bot-ChatSummarizer/master/meme/" +
		filename
}
