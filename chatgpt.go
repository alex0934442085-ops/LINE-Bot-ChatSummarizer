package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// gptGPT3CompleteContext: Call GPT-4o mini API
func gptGPT3CompleteContext(ori string) (ret string) {
	fmt.Printf("Using GPT-4o mini Complete")
	return gptCompleteContext(ori, "gpt-4o-mini")
}

// gptGPT4CompleteContext: Call GPT-4o mini API
func gptGPT4CompleteContext(ori string) (ret string) {
	fmt.Printf("Using GPT-4o mini Complete")
	return gptCompleteContext(ori, "gpt-4o-mini")
}

func gptCompleteContext(ori string, model string) (ret string) {
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: ori,
			},
		},
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		ret = fmt.Sprintf("Err: %v", err)
	} else if len(resp.Choices) == 0 {
		ret = "Err: OpenAI returned no choices"
	} else {
		ret = resp.Choices[0].Message.Content
	}

	return ret
}

// Create image by DALL-E 2
func gptImageCreate(prompt string) (string, error) {
	ctx := context.Background()

	reqURL := openai.ImageRequest{
		Prompt:         prompt,
		Size:           openai.CreateImageSize512x512,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	respURL, err := client.CreateImage(ctx, reqURL)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return "", errors.New("Image creation error")
	}

	if len(respURL.Data) == 0 {
		return "", errors.New("Image creation returned no data")
	}

	fmt.Println(respURL.Data[0].URL)

	return respURL.Data[0].URL, nil
}
