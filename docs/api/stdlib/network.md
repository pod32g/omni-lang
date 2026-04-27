# std.network - Networking Utilities

The `std.network` module provides IP/URL parsing and validation, an HTTP client, raw sockets, and DNS lookups. The pure-data parts (IP/URL validation, status-code categorization, request/response struct construction) work fully offline and on both backends. The actual network I/O (HTTP, DNS, sockets) is provided by the C runtime; running it from a hermetic test environment requires either a stub server or skipping those tests.

## Types

```omni
struct IPAddress {
    address: string
    is_ipv4: bool
    is_ipv6: bool
}

struct URL {
    scheme: string
    host: string
    port: int
    path: string
    query: string
    fragment: string
}

struct HTTPRequest {
    method: string
    url: string
    headers: map<string, string>
    body: string
}

struct HTTPResponse {
    status_code: int
    status_text: string
    headers: map<string, string>
    body: string
}
```

## IP addresses

| Function | Description |
|----------|-------------|
| `ip_parse(ip_str:string): IPAddress` | Wrap a string in an `IPAddress`. Sets `is_ipv4` if the string contains `.` and no `:`, `is_ipv6` if it contains `:`. |
| `ip_is_valid(ip_str:string): bool` | Strict validation. IPv4: four dot-separated decimal segments, each `0..255`. IPv6: colon-separated hex groups (1–4 chars each, ≤8 groups, at most one `::` elision, optional embedded IPv4 tail). |
| `ip_is_loopback(ip:IPAddress): bool` | True for IPv4 addresses starting with `127.`. |
| `ip_is_private(ip:IPAddress): bool` | True for `10.x.x.x`, `172.16.x.x`–`172.31.x.x`, or `192.168.x.x`. |
| `ip_to_string(ip:IPAddress): string` | Returns the original string. |

## URLs

| Function | Description |
|----------|-------------|
| `url_parse(url_str:string): URL` | Parses scheme, host, port, path, query, fragment. |
| `url_to_string(url:URL): string` | Reconstructs the canonical URL string, omitting the default port for `http`/`https`. |
| `url_is_valid(url_str:string): bool` | Requires a scheme matching `[A-Za-z][A-Za-z0-9+.-]*`, then `://`, then a non-empty host before any of `/?#`. Whitespace anywhere is rejected. |

## HTTP client

| Function | Description |
|----------|-------------|
| `http_get(url:string): HTTPResponse` | GET request. Uses libcurl when available, raw socket fallback otherwise. |
| `http_post(url:string, body:string): HTTPResponse` | POST request. |
| `http_put(url:string, body:string): HTTPResponse` | PUT request. |
| `http_delete(url:string): HTTPResponse` | DELETE request. |
| `http_request(req:HTTPRequest): HTTPResponse` | Custom request with headers and body. |

### HTTPResponse helpers

| Function | Description |
|----------|-------------|
| `http_response_is_success(resp): bool` | True for 2xx status codes. |
| `http_response_is_client_error(resp): bool` | True for 4xx. |
| `http_response_is_server_error(resp): bool` | True for 5xx. |
| `http_response_get_header(resp, name): string` | Lookup in headers map. |
| `http_response_set_header(resp, name, value): HTTPResponse` | Mutates and returns the response. |

### HTTPRequest helpers

| Function | Description |
|----------|-------------|
| `http_request_create(method, url): HTTPRequest` | Construct a request with empty headers and body. |
| `http_request_set_header(req, name, value): HTTPRequest` | Set a header. |
| `http_request_set_body(req, body): HTTPRequest` | Set the body. |
| `http_request_get_header(req, name): string` | Lookup a header. |

## DNS

| Function | Description |
|----------|-------------|
| `dns_lookup(hostname:string): array<IPAddress>` | Resolves via `getaddrinfo` (IPv4 and IPv6). |
| `dns_reverse_lookup(ip:IPAddress): string` | Reverse lookup via `getnameinfo`. |

## Sockets

| Function | Description |
|----------|-------------|
| `socket_create(): int` | Create a TCP socket; returns a handle or `-1` on error. |
| `socket_connect(sock, address, port): bool` | Connect to an `address:port`. |
| `socket_bind(sock, address, port): bool` | Bind to an `address:port`. |
| `socket_listen(sock, backlog): bool` | Start listening. |
| `socket_accept(sock): int` | Accept a connection; returns a new handle or `-1`. |
| `socket_send(sock, data): int` | Send bytes; returns count or `-1`. |
| `socket_receive(sock, buffer_size): string` | Receive up to `buffer_size` bytes. |
| `socket_close(sock): bool` | Close the socket. |

## Network info

| Function | Description |
|----------|-------------|
| `network_is_connected(): bool` | Platform-specific reachability check. |
| `network_get_local_ip(): IPAddress` | First non-loopback IPv4 address. |
| `network_ping(host:string): bool` | ICMP on Windows, TCP fallback on POSIX. |

## Constants

`HTTP_OK`, `HTTP_CREATED`, `HTTP_ACCEPTED`, `HTTP_NO_CONTENT`, `HTTP_BAD_REQUEST`, `HTTP_UNAUTHORIZED`, `HTTP_FORBIDDEN`, `HTTP_NOT_FOUND`, `HTTP_METHOD_NOT_ALLOWED`, `HTTP_INTERNAL_SERVER_ERROR`, `HTTP_NOT_IMPLEMENTED`, `HTTP_BAD_GATEWAY`, `HTTP_SERVICE_UNAVAILABLE`.

`HTTP_GET`, `HTTP_POST`, `HTTP_PUT`, `HTTP_DELETE`, `HTTP_HEAD`, `HTTP_OPTIONS`, `HTTP_PATCH`.

`PORT_HTTP` (80), `PORT_HTTPS` (443), `PORT_FTP`, `PORT_SSH`, `PORT_TELNET`, `PORT_SMTP`, `PORT_DNS`, `PORT_POP3`, `PORT_IMAP`, `PORT_HTTPS_ALT` (8080).

## Backend status

The offline pure-data parts (`ip_is_valid`, `ip_parse`, `ip_is_loopback`, `ip_is_private`, `ip_to_string`, `url_is_valid`) are pinned by `TestStdNetworkBasic` on both `omnir` (VM) and `omnic` (C). The audit found:

- `omni_ip_is_valid` previously accepted IPv4 strings with out-of-range segments (e.g. `999.999.999.999`). Now does strict per-segment validation and a real IPv6 check.
- `omni_url_is_valid` previously returned true for any string containing `://`. Now requires a real scheme grammar and a non-empty host.

Network-touching functions (`http_*`, `dns_*`, `socket_*`, `network_ping`, `network_is_connected`, `network_get_local_ip`) are wired but not exercised in CI — running them requires real network or a fixture server.

The C backend currently has gaps in struct field access for `URL` (lowering `omni_url_t*` fields produces type-mismatch warnings/errors) and missing declarations for the OmniLang-defined `http_response_is_*` helpers when used directly. These are tracked as a separate issue and not blocking for the offline audit.
