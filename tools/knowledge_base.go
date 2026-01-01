package tools

import (
	"context"
	"diabetes-agent-mcp-server/config"
	"diabetes-agent-mcp-server/dao"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	client "github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	baseURL            = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	embeddingModelName = "text-embedding-v4"
	collectionName     = "knowledge_doc"
)

var embedder embeddings.Embedder

type VectorDBSearchResult struct {
	Chunk string  `json:"chunk"`
	Score float32 `json:"score"`
}

func init() {
	client, err := openai.New(
		openai.WithEmbeddingModel(embeddingModelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(baseURL),
		openai.WithHTTPClient(&http.Client{
			Timeout: 60 * time.Second,
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create embedder client: %v", err))
	}

	embedder, err = embeddings.NewEmbedder(client)
	if err != nil {
		panic(fmt.Sprintf("Failed to create embedder: %v", err))
	}
}

// SearchUserKnowledgeBase 检索用户知识库，召回用户的知识文件切片
func SearchUserKnowledgeBase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		content := mcp.TextContent{
			Text: "required query param",
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{content},
			IsError: true,
		}, nil
	}

	limit := req.GetInt("limit", defaultSearchResultLimit)

	results, err := retrieveSimilarDocuments(ctx, query, limit)
	if err != nil {
		slog.Error("Failed to search vector store", "err", err)
	}

	slog.Debug("searched user knowledge base", "results", results)

	return mcp.NewToolResultJSON(results)
}

func retrieveSimilarDocuments(ctx context.Context, query string, limit int) ([]VectorDBSearchResult, error) {
	vector, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %v", err)
	}

	// 从上下文获取用户邮箱
	userEmail := ctx.Value("user_email").(string)

	searchOption := client.NewSearchOption(collectionName, limit, []entity.Vector{entity.FloatVector(vector)}).
		WithOutputFields("text").
		WithFilter("user_email == '" + userEmail + "'")

	resultSets, err := dao.MilvusClient.Search(ctx, searchOption)
	if err != nil {
		return nil, fmt.Errorf("error searching vector store: %v", err)
	}

	structedResults := make([]VectorDBSearchResult, 0)
	for _, resSet := range resultSets {
		for i := 0; i < resSet.ResultCount; i++ {
			var text string
			if textColumn := resSet.GetColumn("text"); textColumn != nil {
				if content, ok := textColumn.(*column.ColumnVarChar); ok {
					if content.Len() > 0 {
						text, _ = content.GetAsString(i)
					}
				}
			}

			structedResults = append(structedResults, VectorDBSearchResult{
				Chunk: text,
				Score: resSet.Scores[i],
			})
		}
	}

	return structedResults, nil
}
