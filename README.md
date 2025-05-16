### Commands
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