package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// ============================================================
// Meme 功能
//
// 梗圖
// → 隨機抽一張，不會連續重複，全部抽完才重新洗牌
//
// 梗圖 加班
// → 從檔名包含「加班」的圖片中抽取
//
// 梗圖清單
// → 顯示目前有哪些梗圖
//
// 完全不使用 AI，不消耗 Gemini Token。
// ============================================================

// ============================================================
// 梗圖隨機狀態
//
// key:
// ""      = 全部梗圖
// "加班"  = 加班相關梗圖
// "傻眼"  = 傻眼相關梗圖
//
// 每個關鍵字都有自己的抽取順序。
// ============================================================

var memeState = make(map[string][]string)
var memeMutex sync.Mutex

// ============================================================
// 取得下一張梗圖
//
// 使用「洗牌袋」概念：
//
// 例如有 A B C D
//
// 第一次：C
// 第二次：A
// 第三次：D
// 第四次：B
//
// 四張全部抽過後才重新洗牌。
// ============================================================

func getRandomMeme(images []string, keyword string) string {

	memeMutex.Lock()
	defer memeMutex.Unlock()

	// 如果目前沒有剩餘圖片
	// 或圖片數量發生變化
	// 就重新建立抽取池
	if len(memeState[keyword]) == 0 ||
		!sameImages(memeState[keyword], images) {

		memeState[keyword] = append(
			[]string{},
			images...,
		)

		shuffleImages(
			memeState[keyword],
		)
	}

	// 取出最後一張
	lastIndex := len(memeState[keyword]) - 1

	selected := memeState[keyword][lastIndex]

	// 從抽取池移除
	memeState[keyword] = memeState[keyword][:lastIndex]

	return selected
}

// ============================================================
// 判斷兩個圖片清單是否相同
// ============================================================

func sameImages(
	a []string,
	b []string,
) bool {

	if len(a) != len(b) {
		return false
	}

	aCopy := append([]string{}, a...)
	bCopy := append([]string{}, b...)

	sort.Strings(aCopy)
	sort.Strings(bCopy)

	for i := range aCopy {

		if aCopy[i] != bCopy[i] {
			return false
		}
	}

	return true
}

// ============================================================
// 洗牌
// ============================================================

func shuffleImages(images []string) {

	rand.Shuffle(
		len(images),
		func(i, j int) {
			images[i], images[j] =
				images[j], images[i]
		},
	)
}

// ============================================================
// 梗圖
// ============================================================

func handleMeme(
	event *linebot.Event,
	keyword string,
) {

	keyword = strings.TrimSpace(keyword)

	// ========================================================
	// 讀取 meme 資料夾
	// ========================================================

	memeDir := "meme"

	files, err := os.ReadDir(memeDir)

	if err != nil {

		log.Println(
			"Read meme directory error:",
			err,
		)

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

		// 支援的圖片格式
		if ext != ".jpg" &&
			ext != ".jpeg" &&
			ext != ".png" &&
			ext != ".webp" {

			continue
		}

		// 有指定關鍵字
		// 就只找檔名包含關鍵字的圖片
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
	// 不重複隨機抽取
	// ========================================================

	selected := getRandomMeme(
		images,
		keyword,
	)

	// ========================================================
	// 建立 GitHub Raw 圖片網址
	// ========================================================

	imageURL := buildMemeURL(
		selected,
	)

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
// 梗圖清單
// ============================================================

func handleMemeList(
	event *linebot.Event,
) {

	memeDir := "meme"

	files, err := os.ReadDir(
		memeDir,
	)

	if err != nil {

		log.Println(
			"Read meme directory error:",
			err,
		)

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"柴柴目前找不到梗圖資料夾，嗷……🐕",
			),
		).Do(); err != nil {

			log.Print(err)
		}

		return
	}

	// ========================================================
	// 收集圖片
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

		if ext != ".jpg" &&
			ext != ".jpeg" &&
			ext != ".png" &&
			ext != ".webp" {

			continue
		}

		images = append(
			images,
			filename,
		)
	}

	// ========================================================
	// 沒有梗圖
	// ========================================================

	if len(images) == 0 {

		if _, err := bot.ReplyMessage(
			event.ReplyToken,
			linebot.NewTextMessage(
				"柴柴的梗圖倉庫目前是空的……🐕",
			),
		).Do(); err != nil {

			log.Print(err)
		}

		return
	}

	// ========================================================
	// 排序
	// ========================================================

	sort.Strings(images)

	// ========================================================
	// 建立清單
	// ========================================================

	reply := "🐕 柴柴目前有這些梗圖：\n\n"

	for i, filename := range images {

		reply += fmt.Sprintf(
			"%d. %s\n",
			i+1,
			filename,
		)
	}

	reply += "\n嗷～輸入「梗圖」可以隨機抽一張！"

	// ========================================================
	// 回覆
	// ========================================================

	if _, err := bot.ReplyMessage(
		event.ReplyToken,
		linebot.NewTextMessage(reply),
	).Do(); err != nil {

		log.Print(err)
	}
}

// ============================================================
// 建立 GitHub Raw URL
// ============================================================

func buildMemeURL(
	filename string,
) string {

	return "https://raw.githubusercontent.com/alex0934442085-ops/LINE-Bot-ChatSummarizer/master/meme/" +
		url.PathEscape(filename)
}
