from typing import List

from mcp.server.fastmcp import FastMCP, Image

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

@mcp.tool()
def return_image() -> Image:
    """Return a simple image as bytes"""
    with open("example-image.png", "rb") as img_file:
        return Image(data=img_file.read(), format="png")


if __name__ == "__main__":
    # Start the server
    mcp.run(transport="streamable-http")