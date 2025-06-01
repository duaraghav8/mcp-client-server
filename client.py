from mcp.client.streamable_http import streamablehttp_client
from mcp import ClientSession


async def main():
    # Connect to a streamable HTTP server
    async with streamablehttp_client("http://127.0.0.1:8080/mcp") as (
        read_stream,
        write_stream,
        _,
    ):
        # Create a session using the client streams
        async with ClientSession(read_stream, write_stream) as session:
            # Initialize the connection
            await session.initialize()
            # Call a tool
            tool_result = await session.call_tool("calculator/multiply", {"a": 7, "b": 10})
            print("Result received:")
            print(tool_result.content[0].text)

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
