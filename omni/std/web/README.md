# std.web - HTTP Server Framework

The `std.web` module provides a comprehensive HTTP server framework similar to Express.js or Gin, with support for routing, middleware, async handlers, and advanced features.

## Features

- **Routing**: Support for GET, POST, PUT, DELETE, PATCH, and custom HTTP methods
- **Path Parameters**: Extract parameters from URL paths (e.g., `/user/:id`)
- **Query Parameters**: Parse and access query string parameters
- **Middleware**: Global, route-specific, and group middleware support
- **Async Handlers**: Support for both synchronous and asynchronous request handlers
- **JSON Support**: Built-in JSON parsing and stringifying
- **Form Data**: Parse URL-encoded and multipart form data
- **File Uploads**: Handle file uploads with validation
- **Static Files**: Serve static files with automatic MIME type detection
- **Template Rendering**: Simple template engine with conditionals and loops
- **WebSocket Support**: Full WebSocket support for real-time communication
- **Validation**: Request validation and sanitization utilities
- **Sessions**: Session management with multiple storage backends
- **Authentication**: Password hashing, token generation, and permission checking
- **Rate Limiting**: Configurable rate limiting per client
- **HTTPS/TLS**: Support for secure connections
- **Connection Pooling**: Efficient connection management
- **Graceful Shutdown**: Clean server shutdown with timeout

## Quick Start

```omni
import std

func main():int {
    // Create server
    let server = std.web.server_create(8080, {})
    
    // Add middleware
    std.web.server_use(server, std.web.middleware_logger)
    
    // Define routes
    std.web.server_get(server, "/", |ctx:std.web.Context| {
        return std.web.context_text(ctx, "Hello, World!")
    })
    
    std.web.server_get(server, "/user/:id", |ctx:std.web.Context| {
        let id = std.web.context_param(ctx, "id")
        return std.web.context_text(ctx, "User ID: " + id)
    })
    
    // Start server
    std.web.server_listen(server)
    return 0
}
```

## Server Creation

### `server_create(port:int, options:map<string, any>):Server`

Creates a new HTTP server instance.

**Parameters:**
- `port`: Port number to listen on (1-65535)
- `options`: Configuration options (optional)

**Example:**
```omni
let server = std.web.server_create(8080, {
    "max_connections": 100,
    "timeout": 30
})
```

### `server_listen(server:Server):bool`

Starts the server listening on the configured port.

**Returns:** `true` if server started successfully, `false` otherwise.

### `server_listen_tls(server:Server, cert_file:string, key_file:string):bool`

Starts an HTTPS server with TLS encryption.

**Parameters:**
- `cert_file`: Path to SSL certificate file
- `key_file`: Path to SSL private key file

### `server_close(server:Server)`

Stops the server and cleans up resources.

### `server_graceful_shutdown(server:Server, timeout:int)`

Gracefully shuts down the server, waiting for existing requests to complete.

**Parameters:**
- `timeout`: Maximum seconds to wait for requests to complete

## Routing

### HTTP Methods

```omni
std.web.server_get(server, "/path", handler)
std.web.server_post(server, "/path", handler)
std.web.server_put(server, "/path", handler)
std.web.server_delete(server, "/path", handler)
std.web.server_patch(server, "/path", handler)
std.web.server_all(server, "/path", handler)  // All methods
std.web.server_route(server, "CUSTOM", "/path", handler)  // Custom method
```

### Path Parameters

Extract parameters from URL paths:

```omni
std.web.server_get(server, "/user/:id", |ctx:std.web.Context| {
    let id = std.web.context_param(ctx, "id")
    return std.web.context_text(ctx, "User: " + id)
})
```

**Supported patterns:**
- `:param` - Required parameter
- `*` - Wildcard (matches rest of path)
- `?` - Optional segment

### Route Groups

Group routes with a common prefix and middleware:

```omni
let api = std.web.server_group(server, "/api/v1")
std.web.group_get(api, "/users", handler)
std.web.group_post(api, "/users", handler)
```

## Context API

The `Context` struct provides access to request and response data.

### Request Data

```omni
// Path parameters
let id = std.web.context_param(ctx, "id")

// Query parameters
let query = std.web.context_query(ctx, "q")
let all_queries = std.web.context_query_all(ctx)

// Headers
let user_agent = std.web.context_header(ctx, "User-Agent")

// Body
let body = std.web.context_body(ctx)
let json_data = std.web.context_body_json(ctx)
let form_data = std.web.context_body_form(ctx)

// Cookies
let session_id = std.web.context_get_cookie(ctx, "session_id")
```

### Response Building

```omni
// Text response
std.web.context_text(ctx, "Hello, World!")

// JSON response
std.web.context_json(ctx, {"message": "Success", "data": data})

// HTML response
std.web.context_html(ctx, "<html>...</html>")

// File response
std.web.context_file(ctx, "/path/to/file.html")

// Redirect
std.web.context_redirect(ctx, "/new-location", 302)

// Set headers
std.web.context_set_header(ctx, "Content-Type", "application/json")
std.web.context_status(ctx, 200)

// Set cookies
std.web.context_cookie(ctx, "session_id", "abc123", {
    "max_age": 3600,
    "path": "/",
    "http_only": true
})
```

## Middleware

### Global Middleware

```omni
std.web.server_use(server, std.web.middleware_logger)
std.web.server_use(server, std.web.middleware_cors)
```

### Built-in Middleware

- `middleware_logger` - Log all requests
- `middleware_cors` - Add CORS headers
- `middleware_json_parser` - Parse JSON request bodies
- `middleware_form_parser` - Parse form request bodies
- `middleware_multipart_parser` - Parse multipart form data
- `middleware_static` - Serve static files
- `middleware_compression` - Compress responses with gzip
- `middleware_rate_limit` - Rate limit requests
- `middleware_auth` - Authentication middleware
- `middleware_authorize` - Authorization middleware
- `middleware_error_handler` - Error handling middleware
- `middleware_timeout` - Request timeout middleware
- `middleware_request_size` - Limit request body size
- `middleware_session` - Session management middleware

### Custom Middleware

```omni
func my_middleware(ctx:std.web.Context):std.web.Context {
    // Do something before request
    std.io.println("Processing request...")
    
    // Call next middleware/handler
    // (In actual implementation, this would be handled by the framework)
    
    return ctx
}

std.web.server_use(server, my_middleware)
```

## Async Handlers

Both synchronous and asynchronous handlers are supported:

```omni
// Synchronous handler
std.web.server_get(server, "/sync", |ctx:std.web.Context| {
    return std.web.context_text(ctx, "Synchronous response")
})

// Asynchronous handler
std.web.server_get(server, "/async", async |ctx:std.web.Context| {
    // Perform async operation
    let data = await fetch_data()
    return std.web.context_json(ctx, {"data": data})
})
```

## JSON Support

```omni
// Parse JSON request body
let data = std.web.context_body_json(ctx)

// Send JSON response
std.web.context_json(ctx, {
    "status": "success",
    "data": {"id": 123, "name": "John"}
})
```

## Form Data and File Uploads

### URL-Encoded Forms

```omni
std.web.server_post(server, "/submit", |ctx:std.web.Context| {
    let form = std.web.context_body_form(ctx)
    let name = std.collections.get(form, "name")
    let email = std.collections.get(form, "email")
    return std.web.context_text(ctx, "Received: " + name + ", " + email)
})
```

### Multipart Forms with File Uploads

```omni
std.web.server_post(server, "/upload", |ctx:std.web.Context| {
    let files = std.web.context_files(ctx)
    // Process uploaded files
    return std.web.context_text(ctx, "Files uploaded")
})
```

## Static File Serving

```omni
// Serve static files from a directory
std.web.server_use(server, std.web.middleware_static(server, "/static", "./public", {}))
```

## Template Rendering

```omni
let template = "<h1>{{title}}</h1><p>{{content}}</p>"
let data = {"title": "Hello", "content": "World"}
let html = std.web.template_render(template, data)
std.web.context_html(ctx, html)
```

**Template Syntax:**
- `{{variable}}` - Variable substitution
- `{{#if condition}}...{{/if}}` - Conditionals
- `{{#each items}}...{{/each}}` - Loops
- `{{#with object}}...{{/with}}` - Context switching
- `{{#unless condition}}...{{/unless}}` - Negative conditionals

## WebSocket Support

```omni
std.web.server_websocket(server, "/ws", |ctx:std.web.Context, ws:std.web.WebSocket| {
    // Handle WebSocket connection
    while true {
        let message = std.web.websocket_receive(ws)
        std.web.websocket_send(ws, "Echo: " + message)
    }
})
```

## Validation and Sanitization

```omni
// Validate input
if !std.web.validate_email(email) {
    return std.web.context_status(ctx, 400)
}

// Sanitize output
let safe_html = std.web.sanitize_html(user_input)
let safe_sql = std.web.sanitize_sql(user_query)
```

## Sessions

```omni
// Create session store
let store = std.web.omni_session_store_create("memory")

// In middleware or handler
let session = std.web.omni_session_create(nil, 3600)
std.web.omni_session_set(session, "user_id", "123")
std.web.omni_session_store_save(store, session)
```

## Authentication

```omni
// Hash password
let hash = std.web.omni_auth_hash_password(password, salt)

// Verify password
if std.web.omni_auth_verify_password(password, hash) {
    // Login successful
}

// Generate token
let token = std.web.omni_auth_generate_token(user_id, secret, 3600)

// Verify token
let user_id = std.web.omni_auth_verify_token(token, secret)
```

## Rate Limiting

```omni
let limiter = std.web.omni_rate_limit_create(100, 60)  // 100 requests per 60 seconds

// In middleware
if !std.web.omni_rate_limit_check(limiter, client_ip) {
    return std.web.context_status(ctx, 429)  // Too Many Requests
}
```

## Examples

See `omni/examples/web_server.omni` for a comprehensive example demonstrating all features.

## Implementation Status

All core features are implemented in the runtime. Some advanced features (HTTPS/TLS, full session storage backends, production-grade password hashing) use simplified implementations that can be enhanced for production use.

