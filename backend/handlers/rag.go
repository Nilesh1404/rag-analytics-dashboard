package handlers

import (
	"analytics-backend/db"
	"analytics-backend/llm"
	"analytics-backend/prompts"
	"analytics-backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
 

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func RagHandler(c *gin.Context) {

	var body map[string]string
	c.BindJSON(&body)

	query := body["query"]
    fmt.Println("\nUSER:", query)

	intentRaw, _ := llm.AskLLM(prompts.IntentPrompt + query)
	intentJSON := utils.ExtractJSON(intentRaw)

	var intent map[string]string
	json.Unmarshal([]byte(intentJSON), &intent)

	if intent["type"] == "chat" {
		c.JSON(200, gin.H{"summary": "Ask analytics questions only 🙂"})
		return
	}

	rawPipe, _ := llm.AskLLM(prompts.PipelinePrompt + query)
	pipelineStr := utils.ExtractJSON(rawPipe)
    fmt.Println("\nPIPELINE:", pipelineStr)

	var stages []bson.M
	json.Unmarshal([]byte(pipelineStr), &stages)

	pipeline := mongo.Pipeline{}

	for _, s := range stages {
		for k, v := range s {
			pipeline = append(pipeline, bson.D{{Key: k, Value: v}})
		}
	}

	if strings.Contains(strings.ToLower(query), "last one year") {

		dateStage := bson.D{{
			Key: "$match",
			Value: bson.M{
				"date": bson.M{"$gte": time.Now().AddDate(-1, 0, 0)},
			},
		}}

		pipeline = append(mongo.Pipeline{dateStage}, pipeline...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, _ := db.SalesCollection.Aggregate(ctx, pipeline)

	var results []bson.M
	cursor.All(ctx, &results)
    fmt.Println("\nMONGO RESULTS:", results)

	rawSummary, _ := llm.AskLLM(prompts.SummaryPrompt + fmt.Sprintf("%v", results))
	sumJSON := utils.ExtractJSON(rawSummary)

	var meta map[string]string
	json.Unmarshal([]byte(sumJSON), &meta)
    fmt.Println("\nSUMMARY META:", meta)
       image := ""

     if meta["imagePrompt"] != "" {
	image = llm.GenerateImage(meta["imagePrompt"])
     }

     c.JSON(200, gin.H{
	"summary": meta["summary"],
	"chart":   meta["chart"],
	"image":   image,
	"data":    results,
})

}
