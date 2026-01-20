package tools

import (
	"context"
	"diabetes-agent-mcp-server/dao"
	"fmt"
	"log/slog"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	Neo4jFulltextIndexName = "fulltext_index_entity_name"

	defaultSearchResultLimit = 20
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

// SearchDiabetesKnowledgeGraph 检索基于 DiaKG 构建的图数据库
func SearchDiabetesKnowledgeGraph(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Text: "query param is required",
				},
			},
			IsError: true,
		}, nil
	}

	keywords := strings.Split(query, " ")
	limit := req.GetInt("limit", defaultSearchResultLimit)

	results, err := executeFulltextSearch(ctx, keywords, limit)
	if err != nil {
		slog.Error("Failed to search knowledge graph", "err", err)
	}

	// slog.Debug("searched knowledge graph", "results", results)

	return mcp.NewToolResultJSON(results)
}

// 执行全文搜索，根据 keywords 模糊匹配 Entity 节点的 name 属性
func executeFulltextSearch(ctx context.Context, keywords []string, limit int) ([]KnowlegeGraphSearchResult, error) {
	session := dao.Driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	keywords = cleanKeywords(keywords)
	if len(keywords) == 0 {
		return nil, fmt.Errorf("valid keywords not found")
	}

	// 构建模糊查询条件
	query := strings.Join(keywords, " OR ")

	// 返回匹配查询且至少存在一个关系的节点
	cypherQuery := `
        CALL db.index.fulltext.queryNodes($indexName, $query) 
        YIELD node, score
        WHERE 'Entity' IN labels(node)
        MATCH (node)-[r]-(related:Entity)
        WITH node, score, collect({
            type: type(r),
            related: related {.name, .type}
        }) AS relationships
        RETURN 
            node {.name, .type} AS node,
            relationships,
            score
        ORDER BY score DESC
        LIMIT $limit
    `

	result, err := session.Run(ctx, cypherQuery, map[string]any{
		"indexName": Neo4jFulltextIndexName,
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

func cleanKeywords(keywords []string) []string {
	var escapedKeywords []string
	for _, k := range keywords {
		if k == "" {
			continue
		}
		cleanK := sanitizeLuceneQuery(k)
		if cleanK != "" {
			escapedKeywords = append(escapedKeywords, cleanK)
		}
	}
	return escapedKeywords
}

// 转义 Lucene 保留字符
func sanitizeLuceneQuery(query string) string {
	replacer := strings.NewReplacer(
		"+", "\\+", "-", "\\-", "&", "\\&", "|", "\\|",
		"!", "\\!", "(", "\\(", ")", "\\)", "{", "\\{",
		"}", "\\}", "[", "\\[", "]", "\\]", "^", "\\^",
		"\"", "\\\"", "~", "\\~", "*", "\\*", "?", "\\?",
		":", "\\:", "\\", "\\\\", "/", "\\/",
	)
	return replacer.Replace(query)
}
