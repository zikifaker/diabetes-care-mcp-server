package tool

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	DefaultSearchResultLimit = 20
	Neo4jFulltextName        = "fulltextSearch"
)

type KnowlegeGraphSearchResult struct {
	Node          EntityNode `json:"node"`
	Relationships []Relation `json:"relationships"`
	Score         float64    `json:"score"`
}

type EntityNode struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Relation struct {
	Type    string     `json:"type"`
	Related EntityNode `json:"related"`
}

type VectorDBSearchResult struct {
}

// SearchDiabetesKnowledgeBase 检索糖尿病知识库
// 先分别进行图检索和向量检索，各自召回 limit / 2 条
func SearchDiabetesKnowledgeBase(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		content := mcp.TextContent{
			Text: "query param is required",
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{content},
			IsError: true,
		}, nil
	}

	limit := req.GetInt("limit", DefaultSearchResultLimit)
	if limit <= 0 {
		limit = DefaultSearchResultLimit
	}

	knowledgeGraphResults, err := searchKnowledgeGraph(ctx, query, limit/2)
	if err != nil {
		slog.Error("Failed to search knowledge graph", "err", err)
	}

	vectorStoreResults, err := searchVectorStore(ctx, query, limit/2)
	if err != nil {
		slog.Error("Failed to search vector store", "err", err)
	}

	finalResults := map[string]any{
		"knowledge_graph_results": knowledgeGraphResults,
		"vector_store_results":    vectorStoreResults,
	}

	slog.Debug("search diabetes knowledge base finished", "final_results", finalResults)

	return mcp.NewToolResultJSON(finalResults)
}

// 检索图数据库（DiaKG数据集）
func searchKnowledgeGraph(ctx context.Context, query string, limit int) ([]KnowlegeGraphSearchResult, error) {
	session := dao.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	// 查询至少有一个关系的节点
	cypherQuery := `
		CALL db.index.fulltext.queryNodes($indexName, $query) 
		YIELD node, score
		WHERE 'Entity' IN labels(node)
		WITH node, score, [(node)-[r]-(related:Entity) | {
			type: type(r),
			related: related {.name, .type}
		}] AS relationships
		WHERE size(relationships) > 0
		RETURN 
			node {.name, .type} AS node,
			relationships,
			score
		ORDER BY score DESC
		LIMIT $limit
	`

	result, err := session.Run(ctx, cypherQuery, map[string]any{
		"indexName": Neo4jFulltextName,
		"query":     query,
		"limit":     limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute fulltext query: %v", err)
	}

	// 将 map 转换为结构体
	var results []KnowlegeGraphSearchResult
	for result.Next(ctx) {
		var sr KnowlegeGraphSearchResult
		if err := mapstructure.Decode(result.Record().AsMap(), &sr); err != nil {
			return nil, fmt.Errorf("failed to decode search result: %v", err)
		}
		results = append(results, sr)
	}

	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("failed to process search results: %v", err)
	}

	return results, nil
}

// 检索向量存储（用户上传的知识文件）
func searchVectorStore(ctx context.Context, query string, limit int) ([]map[string]any, error) {
	return nil, nil
}
