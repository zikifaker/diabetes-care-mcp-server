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
				mcp.Max(30),
				mcp.Description("Number of results to return (10-30)"),
			),
		),
		tools.SearchDiabetesKnowledgeGraph,
	)

	s.AddTool(
		mcp.NewTool("fetch_health_data",
			mcp.WithDescription("Get user health data including blood glucose records, health profile, and exercise records."),
			mcp.WithString("type",
				mcp.Required(),
				mcp.Enum("blood_glucose", "health_profile", "exercise_records"),
				mcp.Description("Type of health data to retrieve"),
			),
			mcp.WithNumber("limit",
				mcp.Description("Number of most recent records to return (10-100, only for blood_glucose and exercise_records)"),
				mcp.Min(10),
				mcp.Max(100),
			),
		),
		tools.FetchHealthData,
	)
}
