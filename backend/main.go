package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ollamaURL = "http://host.docker.internal:11434"

type GenerateRequest struct {
	Content   string `json:"content" binding:"required"`
	MaxTokens int    `json:"max_tokens"`
}

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	r.GET("/models", func(c *gin.Context) {
		resp, err := http.Get(ollamaURL + "/api/tags")
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch models"})
			return
		}
		defer resp.Body.Close()

		var tagsResp struct {
			Models []struct {
				Model string `json:"model"`
			} `json:"models"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response from Ollama"})
			return
		}

		models := make([]string, 0)
		for _, m := range tagsResp.Models {
			models = append(models, m.Model)
		}
		c.JSON(http.StatusOK, models)
	})

	r.GET("/generate", func(c *gin.Context) {
		model := c.Query("model")
		content := c.Query("content")
		if model == "" || content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model and content query params required"})
			return
		}

		req := GenerateRequest{
			Content:   content,
			MaxTokens: 200,
		}

		streamResponse(c, model, req)
	})

	r.POST("/generate", func(c *gin.Context) {
		model := c.Query("model")
		if model == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "model query param required"})
			return
		}

		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		streamResponse(c, model, req)
	})

	log.Println("🚀 Backend listening on http://localhost:1111")
	r.Run(":1111")
}

func streamResponse(c *gin.Context, model string, req GenerateRequest) {
	payload := fmt.Sprintf(`{
		"model": "%s",
		"prompt": %q,
		"max_tokens": %d,
		"stream": true
	}`, model, req.Content, req.MaxTokens)

	ollamaReq, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		ollamaURL+"/api/generate",
		strings.NewReader(payload),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	ollamaReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(ollamaReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to reach Ollama"})
		return
	}
	defer resp.Body.Close()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			fmt.Fprintf(c.Writer, "data: %s\n\n", line)
			flusher.Flush()
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("streaming error: %v", err)
	}
}
