# MCP

To run mpc host, you may use (mcphost)[https://github.com/mark3labs/mcphost].
You would need an LLM model, for example [Ollama](https://ollama.com/).

This example uses `mcphost` with the `mistral` model from Ollama.
```
mcphost --config ferretdb-mcp-server/.mcp.json --model ollama:mistral
```

An example of `.mcp.json` configuration file for the MCP server is below:
```
{
  "mcpServers": {
    "server_name": {
      "url": "http://127.0.0.1:8081/sse",
      "headers":[
        "Authorization: Basic <credentials>"
      ]
    }
  }
}
```
