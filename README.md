# MCP Client-Server
The simplest example of MCP client-server architecture.

The server can be run as a remote MCP server serving HTTP requests.

Client consumes the remote server over HTTP.


### Commands
One-time setup
```
# Setup virtualenv
python3 -m venv venv
source venv/bin/activate

# Download all dependencies using uv
uv install
```

---

First, activate the virtual environment:
```
source venv/bin/activate

```

Then proceed with the following:

1. Start mcp server in dev mode to test out tools
```
mcp dev server.py
```

2. Run the MCP server
```
python server.py
```
Accessible at http://0.0.0.0:8000/mcp by default

3. Run the MCP client that calls a tool on server
```
python client.py
```