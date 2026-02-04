package llm

import (
	"analytics-backend/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func GenerateImage(prompt string) string {

	if prompt == "" {
		return ""
	}

	reqBody := map[string]interface{}{
		"model": "gpt-image-1",
		"prompt": prompt,
		"size": "1024x1024",
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/images/generations",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+config.OPENAI_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("IMAGE ERROR:", err)
		return ""
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)

	fmt.Println("\nIMAGE RAW RESPONSE:", raw)

	if raw["data"] == nil {
		fmt.Println("NO IMAGE DATA")
		return ""
	}

	data := raw["data"].([]interface{})
	img := data[0].(map[string]interface{})

	return img["b64_json"].(string)
}
