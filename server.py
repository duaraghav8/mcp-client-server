from typing import List

from mcp.server.fastmcp import FastMCP

# Create an MCP server
mcp = FastMCP("Demo", stateless_http=True)


@mcp.tool(name="myserver/add")
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

@mcp.tool()
def return_list() -> List[str]:
    """Return a list of strings"""
    return ["apple", "banana", "cherry"]

@mcp.tool()
def return_dict() -> dict:
    """Return a dictionary"""
    return {"name": "Alice", "age": 30, "city": "New York"}


if __name__ == "__main__":
    # Start the server
    mcp.run(transport="streamable-http")