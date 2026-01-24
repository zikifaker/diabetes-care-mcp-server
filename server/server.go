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
				mcp.Description("Number of results to return (10-20)"),
			),
		),
		tools.SearchDiabetesKnowledgeGraph,
	)

	s.AddTool(
		mcp.NewTool("get_blood_glucose_records",
			mcp.WithDescription(`
				Retrieve the user's recent blood glucose measurements.
				Returns blood glucose records including blood glucose value, measured time, and dining status at measurement.
			`),
			mcp.WithNumber("limit",
				mcp.Required(),
				mcp.Min(10),
				mcp.Max(100),
				mcp.Description("Number of most recent records to return (10-100)"),
			),
		),
		tools.GetBloodGlucoseRecords,
	)

	s.AddTool(
		mcp.NewTool("get_health_profile",
			mcp.WithDescription(`
				Retrieve the user's comprehensive health profile information.
				Returns the user's diabetes type, medical history, and complication status.
			`),
		),
		tools.GetHealthProfile,
	)
}
