package main

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// gptGPT3CompleteContext:
// 保留原本函式名稱，這樣 bot.go 不需要修改。
func gptGPT3CompleteContext(ori string) (ret string) {
	fmt.Println("Using Gemini 2.5 Flash")
	return geminiCompleteContext(ori)
}

// gptGPT4CompleteContext:
// 目前也使用 Gemini 2.5 Flash。
// 保留原本函式名稱，避免 bot.go 需要修改。
func gptGPT4CompleteContext(ori string) (ret string) {
	fmt.Println("Using Gemini 2.5 Flash")
	return geminiCompleteContext(ori)
}

// geminiCompleteContext: Call Gemini API
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
	defer client.Close()

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(ori),
		nil,
	)

	if err != nil {
		return fmt.Sprintf("Err: Gemini API: %v", err)
	}

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

// 圖片功能目前先保留。
// 因為我們這次主要是要讓 :sum_all 使用免費 Gemini 摘要。
// Gemini 2.5 Flash 本身不是圖片生成模型。
func gptImageCreate(prompt string) (string, error) {
	return "", fmt.Errorf("目前圖片生成功能尚未改成 Gemini")
}
