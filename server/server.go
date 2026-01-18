package server

import (
	"diabetes-agent-mcp-server/middleware"
	"diabetes-agent-mcp-server/tools"
	_ "embed"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "Diabetes Care MCP Server"
	serverVersion = "1.0.0"
)

//go:embed prompts/search_diabetes_kg/query.txt
var searchDiabetesKGQueryDesc string

func NewHTTPServer() *server.StreamableHTTPServer {
	hooks := &server.Hooks{}

	// 注册 hook，推送工具调用结果
	hooks.AddAfterCallTool(pushCallToolResult)

	s := server.NewMCPServer(serverName, serverVersion,
		server.WithToolCapabilities(true),
		server.WithToolHandlerMiddleware(middleware.AuthMiddleware),
		server.WithHooks(hooks),
	)

	registerTools(s)

	return server.NewStreamableHTTPServer(s)
}

func registerTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("search_diabetes_knowledge_graph",
			mcp.WithDescription(`
				Search professional information about diabetes guidelines, medications, diagnostics, and treatments. 
				Returns structured data from knowledge graph (entities and relationships). 
				All results are sorted by relevance score in descending order.
			`),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description(searchDiabetesKGQueryDesc),
			),
			mcp.WithNumber("limit",
				mcp.Min(10),
				mcp.Max(20),
				mcp.Description("Number of results to return"),
			),
		),
		tools.SearchDiabetesKnowledgeGraph,
	)
}
