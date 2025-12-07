package server

import (
	"diabetes-agent-mcp-server/tool"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "Diabetes Agent MCP Server"
	serverVersion = "1.0.0"
)

func NewHTTPServer() *server.StreamableHTTPServer {
	s := server.NewMCPServer(serverName, serverVersion,
		server.WithToolCapabilities(true),
	)

	s.AddTool(
		mcp.NewTool("diabetes_knowledge_base_search",
			mcp.WithDescription(`
				Search professional information about diabetes guidelines, 
				medications, diagnostics, and treatments. Returns structured data 
				from both knowledge graph (entities and relationships) and vector storage 
				(semantic text chunks). All results are sorted by relevance score in descending order.
			`),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("Search query string"),
			),
			mcp.WithNumber("limit",
				mcp.DefaultNumber(tool.DefaultSearchResultLimit),
				mcp.Description("An even number of results to return (recommended between 10-50)"),
			),
		),
		tool.SearchDiabetesKnowledgeBase,
	)

	return server.NewStreamableHTTPServer(s)
}
