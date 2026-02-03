package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var salesCollection *mongo.Collection

// ================= CONFIG =================

const OPENAI_KEY = "api key here"
// ================= PROMPTS =================

const schemaPrompt = `
You know this MongoDB schema.

Collection: sales
Fields:
- productId (string)
- date (date)
- quantity (number)
- revenue (number)
- profit (number)

Collection: products
Fields:
- _id (string)
- name (string)
- category (string)
- price (number)
- profitMargin (number)

Join rule:
sales.productId == products._id

Each sales document represents one transaction on one date.

Revenue and profit analytics are computed by aggregating sales.

Cumulative percentages (70%, 60%) are NOT done in MongoDB.
They are handled by backend code.
`


const intentPrompt = `
You are a classifier.

If the user asks about:
- revenue
- profit
- sales
- products
- quantities
- trends
- top selling

Then return analytics.

Otherwise return chat.

Return STRICT JSON inside <json></json> only.

<json>
{ "type":"analytics|chat" }
</json>

User Question:
`



const pipelinePrompt = `
You are a MongoDB aggregation translator.

Convert English questions into MongoDB aggregation pipelines.

STRICT RULES:

- Return JSON ARRAY inside <json></json>
- Use ONLY double quotes
- NO explanation
- NO markdown
- NO JavaScript
- NO ISODate
- DO NOT calculate percentages
- DO NOT do cumulative logic
- DO NOT filter by date unless user explicitly provides dates
- Backend will handle "last year" or "last 3 years"

Allowed operators:
"$match","$lookup","$unwind","$group","$sum","$sort","$limit"

When grouping products:
Always group by product name.

When matching product name:
Use "$regex" with "$options":"i"

If user asks for:
- top selling → sort by totalRevenue desc
- trend → group by year extracted from date string
- profit → sum profit
- revenue → sum revenue

Example:

<json>
[
 {
  "$lookup":{
   "from":"products",
   "localField":"productId",
   "foreignField":"_id",
   "as":"productInfo"
  }
 },
 { "$unwind":"$productInfo" },
 {
  "$group":{
   "_id":"$productInfo.name",
   "productName":{"$first":"$productInfo.name"},
   "totalRevenue":{"$sum":"$revenue"},
   "totalProfit":{"$sum":"$profit"}
  }
 },
 { "$sort":{"totalRevenue":-1} }
]
</json>

User Question:
`


const summaryPrompt = `
You are a senior analytics architect.

Rules:
- ONLY use provided Mongo results
- NEVER invent numbers
- NEVER change totals
- NEVER add products

You MUST embed ALL product names and numeric values directly into imagePrompt.

Return STRICT JSON inside <json></json>.

Format:

<json>
{
 "summary":"short business explanation using Mongo values",
 "chart":"pie|bar|line",
 "imagePrompt":"Real BI dashboard showing EXACT MongoDB data: list each product with revenue and profit, labeled pie chart with percentages, bar chart with product vs revenue, line chart trend, KPI cards showing totals. MUST include these values explicitly:",
 "videoPrompt":"optional"
}
</json>

Mongo Results:
`





// ================= OPENAI CLIENT =================

func askLLM(prompt string) (string, error) {

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": schemaPrompt},
			{"role": "user", "content": prompt},
		},
		"temperature": 0,
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Authorization", "Bearer "+OPENAI_KEY)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)

	// 🔴 LOG FULL RESPONSE
	fmt.Println("\nOPENAI RAW RESPONSE:")
	fmt.Println(raw)

	// Handle OpenAI errors safely
	if raw["choices"] == nil {
		return "", fmt.Errorf("openai error: %v", raw)
	}

	choices := raw["choices"].([]interface{})
	msg := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	content := msg["content"].(string)

	return content, nil
}


// ================= HELPERS =================

func extractJSON(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	start := strings.Index(s, "<json>")
	end := strings.Index(s, "</json>")
	if start == -1 || end == -1 {
		return ""
	}
	return strings.TrimSpace(s[start+6 : end])
}

// ================= MAIN =================

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}

	salesCollection = client.Database("analyticsDB").Collection("sales")

	fmt.Println("MongoDB Connected")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	r.POST("/rag", ragHandler)

	r.Run(":8080")
}

// ================= RAG HANDLER =================

func ragHandler(c *gin.Context) {

	var body map[string]string
	c.BindJSON(&body)

	query := body["query"]

	fmt.Println("\nUSER:", query)

	// ---------- STEP 1: INTENT (RAG) ----------

	intentRaw, _ := askLLM(intentPrompt + query)
	intentJSON := extractJSON(intentRaw)

	var intent map[string]string
	json.Unmarshal([]byte(intentJSON), &intent)

	// If normal chat → answer directly
	if intent["type"] == "chat" {

	c.JSON(200, gin.H{
		"summary": "🙂 Please ask questions related to product sales analytics (revenue, profit, products, quantities).",
	})

	return
}


	// ---------- STEP 2: LLM → MONGO PIPELINE ----------

	rawPipe, _ := askLLM(pipelinePrompt + query)
	pipelineStr := extractJSON(rawPipe)

	fmt.Println("\nPIPELINE:", pipelineStr)

	var stages []bson.M
	if err := json.Unmarshal([]byte(pipelineStr), &stages); err != nil {
		c.JSON(200, gin.H{"summary": "Invalid Mongo pipeline"})
		return
	}

	pipeline := mongo.Pipeline{}
	for _, s := range stages {
		for k, v := range s {
			pipeline = append(pipeline, bson.D{{Key: k, Value: v}})
		}
	}

	// -------- MANUAL DATE FILTER (server-side, not GPT) --------

lowerQuery := strings.ToLower(query)

if strings.Contains(lowerQuery, "last one year") ||
   strings.Contains(lowerQuery, "past year") {

	oneYearAgo := time.Now().AddDate(-1, 0, 0)

	dateStage := bson.D{{
		Key: "$match",
		Value: bson.M{
			"date": bson.M{
				"$gte": oneYearAgo,
			},
		},
	}}

	// prepend date filter
	pipeline = append(mongo.Pipeline{dateStage}, pipeline...)

	fmt.Println("✅ Injected last-year date filter:", oneYearAgo)
}

	// ---------- STEP 3: MONGO RETRIEVAL ----------

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := salesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(500, err)
		return
	}

	var results []bson.M
	cursor.All(ctx, &results)

	fmt.Println("\nMONGO RESULTS:", results)

	// ---------- STEP 4: LLM ANALYSIS ----------

	rawSummary, _ := askLLM(summaryPrompt + fmt.Sprintf("%v", results))
	sumJSON := extractJSON(rawSummary)

	var meta map[string]string
	json.Unmarshal([]byte(sumJSON), &meta)
	imageURL := ""

if meta["imagePrompt"] != "" {
	imageURL = generateImage(meta["imagePrompt"])
}


	// ---------- RESPONSE ----------

c.JSON(200, gin.H{
	"summary": meta["summary"],
	"chart": meta["chart"],
	"image": imageURL,
	"data": results,
})


}
func generateImage(prompt string) string {

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

	req.Header.Set("Authorization", "Bearer "+OPENAI_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("IMAGE REQUEST ERROR:", err)
		return ""
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&raw)

	fmt.Println("\nIMAGE RAW RESPONSE:", raw)

	// OpenAI error
	if raw["data"] == nil {
		fmt.Println("NO IMAGE DATA")
		return ""
	}

	data, ok := raw["data"].([]interface{})
	if !ok || len(data) == 0 {
		fmt.Println("EMPTY IMAGE ARRAY")
		return ""
	}

	img := data[0].(map[string]interface{})

	b64, ok := img["b64_json"].(string)
	if !ok {
		fmt.Println("NO BASE64 IMAGE")
		return ""
	}

	return b64
}



