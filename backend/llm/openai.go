package llm

import (
	"analytics-backend/config"
	"analytics-backend/prompts"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func AskLLM(prompt string) (string, error) {

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": prompts.SchemaPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0,
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(body))

	req.Header.Set("Authorization", "Bearer "+config.OPENAI_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)
    fmt.Println("\nOPENAI RAW RESPONSE:")
    fmt.Println(raw)

	if raw["choices"] == nil {
		return "", fmt.Errorf("openai error")
	}

	choices := raw["choices"].([]interface{})
	msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})

	return msg["content"].(string), nil
}
