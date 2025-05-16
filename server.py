from mcp.server.fastmcp import FastMCP

# Create an MCP server
mcp = FastMCP("Demo", stateless_http=True)


@mcp.tool()
def add(a: int, b: int) -> int:
    """Add two numbers"""
    return a + b

@mcp.tool()
def subtract(a: int, b: int) -> int:
    """Subtract two numbers"""
    return a - b

@mcp.tool()
def multiply(a: int, b: int) -> int:
    """Multiply two numbers"""
    return a * b


if __name__ == "__main__":
    # Start the server
    mcp.run(transport="streamable-http")