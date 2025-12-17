package server

import (
	"diabetes-agent-mcp-server/middleware"
	"diabetes-agent-mcp-server/tool"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "Diabetes Agent MCP Server"
	serverVersion = "1.0.0"
)

func NewHTTPServer() *server.StreamableHTTPServer {
	hooks := &server.Hooks{}

	// 注册工具调用成功后将调用结果推送给客户端的 hook
	hooks.AddAfterCallTool(pushCallToolResult)

	s := server.NewMCPServer(serverName, serverVersion,
		server.WithToolCapabilities(true),
		server.WithToolHandlerMiddleware(middleware.AuthMiddleware),
		server.WithHooks(hooks),
	)

	addTools(s)

	return server.NewStreamableHTTPServer(s)
}

func addTools(s *server.MCPServer) {
	s.AddTool(
		mcp.NewTool("diabetes_knowledge_graph",
			mcp.WithDescription(`
				Search professional information about diabetes guidelines, medications, diagnostics, and treatments. 
				Returns structured data from knowledge graph (entities and relationships). 
				All results are sorted by relevance score in descending order.
			`),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query string"),
			),
			mcp.WithNumber("limit",
				mcp.DefaultNumber(tool.DefaultSearchResultLimit),
				mcp.Description("An even number of results to return (recommended between 10-20)"),
			),
		),
		tool.SearchDiabetesKnowledgeGraph,
	)

	s.AddTool(
		mcp.NewTool("user_knowledge_base",
			mcp.WithDescription(`
				Search the user's private knowledge base containing personal documents and information across various domains. 
				Use this tool when you need to find specific information from the user's personal knowledge collection, 
				especially when general knowledge is insufficient.`),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query string"),
			),
			mcp.WithNumber("limit",
				mcp.DefaultNumber(tool.DefaultSearchResultLimit),
				mcp.Description("Number of results to return (recommended between 10-20)"),
			),
		),
		tool.SearchUserKnowledgeBase,
	)
}
