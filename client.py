import base64

from mcp.client.streamable_http import streamablehttp_client
from mcp import ClientSession


async def main():
    # Connect to a streamable HTTP server
    async with streamablehttp_client("http://127.0.0.1:8000/mcp") as (
        read_stream,
        write_stream,
        _,
    ):
        # Create a session using the client streams
        async with ClientSession(read_stream, write_stream) as session:
            # Initialize the connection
            init_response = await session.initialize()
            print("Initialization response:", init_response.serverInfo)
            # Call a tool
            tool_result = await session.call_tool("multiply", {"a": 7, "b": 10})
            print("Result received:")
            print(tool_result.content[0].text)

            tool_result = await session.call_tool("return_image")
            print("Image received:")
            print(tool_result.content[0].mimeType, tool_result.content[0].type)
            # print(tool_result.content)
            with open("received_image.png", "wb") as img_file:
                img_file.write(base64.b64decode(tool_result.content[0].data))

if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
