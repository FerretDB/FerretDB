# MCP

To run mpc host, you would need an LLM model such as [Ollama](https://ollama.com/) install on your host.

This example uses the `mistral` model from Ollama, executed at the root of this repository:
```
bin/mcphost --config build/mcp/.mcp.json --model ollama:mistral
```
