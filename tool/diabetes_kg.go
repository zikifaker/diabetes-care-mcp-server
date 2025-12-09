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
	DefaultSearchResultLimit = 10
	Neo4jFulltextName        = "fulltextSearch"
)

type KnowlegeGraphSearchResult struct {
	Node          EntityNode `json:"node"`
	Relationships []Relation `json:"relationships"`
	Score         float32    `json:"score"`
}

type EntityNode struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Relation struct {
	Type    string     `json:"type"`
	Related EntityNode `json:"related"`
}

// SearchDiabetesKnowledgeGraph 检索基于 DiaKG 构建的图数据库。召回实体-关系
func SearchDiabetesKnowledgeGraph(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	results, err := executeFulltextSearch(ctx, query, limit)
	if err != nil {
		slog.Error("Failed to search knowledge graph", "err", err)
	}

	slog.Debug("searched knowledge graph", "results", results)

	return mcp.NewToolResultJSON(results)
}

// 执行全文搜索，根据 query 匹配 Entity 节点的 name 和 type 属性
func executeFulltextSearch(ctx context.Context, query string, limit int) ([]KnowlegeGraphSearchResult, error) {
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
