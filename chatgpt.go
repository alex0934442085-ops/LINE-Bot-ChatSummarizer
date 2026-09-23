package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// gptGPT3CompleteContext:
// 保留原本函式名稱，這樣 bot.go 不需要修改。
func gptGPT3CompleteContext(ori string) (ret string) {
	fmt.Println("Using Gemini 3.6 Flash")
	return geminiCompleteContext(ori)
}

// gptGPT4CompleteContext:
// 保留原本函式名稱，這樣 bot.go 不需要修改。
func gptGPT4CompleteContext(ori string) (ret string) {
	fmt.Println("Using Gemini 3.6 Flash")
	return geminiCompleteContext(ori)
}

// geminiCompleteContext:
// Call Gemini API
//
// 遇到 Gemini 暫時性 503 / UNAVAILABLE 時，
// 自動使用 exponential backoff 重試。
func geminiCompleteContext(ori string) (ret string) {
	ctx := context.Background()

	apiKey := os.Getenv("GeminiApiKey")

	if apiKey == "" {
		return "Err: GeminiApiKey 沒有設定"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return fmt.Sprintf("Err: 無法建立 Gemini Client: %v", err)
	}

	// 最多嘗試 4 次：
	// 第 1 次：立即執行
	// 第 2 次：等待 1 秒
	// 第 3 次：等待 2 秒
	// 第 4 次：等待 4 秒
	const maxAttempts = 4

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {

		// 第一次不用等待
		if attempt > 1 {
			waitSeconds := 1 << (attempt - 2)

			fmt.Printf(
				"Gemini 暫時不可用，%d 秒後進行第 %d/%d 次重試...\n",
				waitSeconds,
				attempt,
				maxAttempts,
			)

			time.Sleep(
				time.Duration(waitSeconds) * time.Second,
			)
		}

		result, err := client.Models.GenerateContent(
			ctx,
			"gemini-3.6-flash",
			genai.Text(ori),
			nil,
		)

		if err != nil {

			lastErr = err

			errText := err.Error()

			// 只有這類暫時性錯誤才重試
			if isGeminiRetryableError(errText) {

				fmt.Printf(
					"Gemini API 暫時錯誤（第 %d/%d 次）：%v\n",
					attempt,
					maxAttempts,
					err,
				)

				continue
			}

			// 非暫時性錯誤，不需要重試
			return fmt.Sprintf(
				"Err: Gemini API: %v",
				err,
			)
		}

		// ========================================================
		// Gemini 沒有錯誤，開始檢查回傳內容
		// ========================================================

		if result == nil || len(result.Candidates) == 0 {
			return "Err: Gemini 沒有回傳結果"
		}

		if result.Candidates[0].Content == nil ||
			len(result.Candidates[0].Content.Parts) == 0 {

			return "Err: Gemini 回傳內容為空"
		}

		for _, part := range result.Candidates[0].Content.Parts {

			if part.Text != "" {
				return part.Text
			}
		}

		return "Err: Gemini 回傳文字為空"
	}

	// ========================================================
	// 所有重試都失敗
	// ========================================================

	return fmt.Sprintf(
		"Err: Gemini API 暫時忙碌，已重試 %d 次，請稍後再試。最後錯誤：%v",
		maxAttempts,
		lastErr,
	)
}

// isGeminiRetryableError:
// 判斷 Gemini 錯誤是否屬於值得重試的暫時性錯誤。
func isGeminiRetryableError(errText string) bool {

	errText = strings.ToUpper(errText)

	// HTTP 503
	if strings.Contains(errText, "503") {
		return true
	}

	// Google API UNAVAILABLE
	if strings.Contains(errText, "UNAVAILABLE") {
		return true
	}

	// 暫時性服務過載
	if strings.Contains(errText, "HIGH DEMAND") {
		return true
	}

	// RESOURCE_EXHAUSTED 有時代表暫時性流量限制
	if strings.Contains(errText, "RESOURCE_EXHAUSTED") {
		return true
	}

	return false
}

// 圖片功能目前先保留。
// 這次主要是讓 :sum_all 使用 Gemini 摘要。
// Gemini 3.6 Flash 本身不是圖片生成模型。
func gptImageCreate(prompt string) (string, error) {
	return "", fmt.Errorf("目前圖片生成功能尚未改成 Gemini")
}
