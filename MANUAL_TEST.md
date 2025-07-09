# Manual Test Instructions

## Simple Manual Test

1. **Start the server:**
   ```bash
   go run .
   ```

2. **Copy and paste this initialize message:**
   ```json
   {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{"tools":true},"clientInfo":{"name":"test-client","version":"1.0.0"}}}
   ```

3. **You should see a response like:**
   ```json
   {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"resources":{"subscribe":true},"tools":{"listChanged":true}},"serverInfo":{"name":"GitLab MCP Server","version":"1.0.0"}}}
   ```

4. **Then copy and paste this tool call:**
   ```json
   {"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project_id","arguments":{"remote_url":"git@gitlab.com:example/example-project.git"}}}
   ```

5. **You should see debug logs on stderr and a response with the project ID**

6. **Press Ctrl+C to stop the server**
