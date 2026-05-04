package vm

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// vmHTTPClient is the shared client for all VM-driven HTTP intrinsics.
// A 30s timeout matches the C runtime's libcurl/socket fallbacks closely
// enough that tests behave the same on both backends.
var vmHTTPClient = &http.Client{Timeout: 30 * time.Second}

// vmHTTPDo issues a request and converts the response to the OmniLang
// HTTPResponse struct shape (map[string]interface{} matching execMember's
// expectations). On any error a status_code-0 stub is returned — same
// failure mode the C runtime uses, so callers can branch on
// is_success / is_client_error / is_server_error uniformly.
func vmHTTPDo(method, url string, headers map[interface{}]interface{}, body string) map[string]interface{} {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return vmHTTPResponseStub()
	}
	for k, v := range headers {
		ks, ok1 := k.(string)
		vs, ok2 := v.(string)
		if !ok1 || !ok2 {
			continue
		}
		req.Header.Set(ks, vs)
	}
	resp, err := vmHTTPClient.Do(req)
	if err != nil {
		return vmHTTPResponseStub()
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return vmHTTPResponseStub()
	}

	respHeaders := make(map[interface{}]interface{}, len(resp.Header))
	for k, v := range resp.Header {
		if len(v) == 0 {
			continue
		}
		// Net/http stores multi-valued headers as []string; the OmniLang
		// surface is single-valued, so keep the first per name (most
		// semantic for HTTP — only Set-Cookie / proxy auth would differ).
		respHeaders[k] = v[0]
	}

	statusText := resp.Status // e.g. "200 OK"
	if idx := strings.IndexByte(statusText, ' '); idx >= 0 {
		statusText = statusText[idx+1:]
	}

	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"status_text": statusText,
		"headers":     respHeaders,
		"body":        string(bodyBytes),
	}
}

// vmHTTPResponseStub mirrors the OmniLang stub HTTPResponse {0, "", {}, ""}
// that the .omni body would have returned on _intrinsic_not_wired. Returning
// this shape (instead of nil) lets callers do .status_code etc. without a
// panic — the same defensive contract the C runtime offers.
func vmHTTPResponseStub() map[string]interface{} {
	return map[string]interface{}{
		"status_code": 0,
		"status_text": "",
		"headers":     map[interface{}]interface{}{},
		"body":        "",
	}
}

// vmHTTPRequestExtract pulls method/url/headers/body out of an HTTPRequest
// struct value (map[string]interface{}). Missing fields fall back to safe
// defaults — GET / empty url / empty headers / empty body — to match the
// stub behaviour of the runtime when called with a partially-built request.
func vmHTTPRequestExtract(reqValue interface{}) (method, url string, headers map[interface{}]interface{}, body string) {
	method = "GET"
	headers = map[interface{}]interface{}{}
	m, ok := reqValue.(map[string]interface{})
	if !ok {
		return
	}
	if v, ok := m["method"].(string); ok && v != "" {
		method = v
	}
	if v, ok := m["url"].(string); ok {
		url = v
	}
	if v, ok := m["headers"].(map[interface{}]interface{}); ok {
		headers = v
	}
	if v, ok := m["body"].(string); ok {
		body = v
	}
	return
}

// vmIntFromAny coerces an int-or-int32 stored in interface{} to int. Used
// by http_response_is_* helpers when reading status_code from a struct
// field (map[string]interface{} can hold either depending on origin).
func vmIntFromAny(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case int32:
		return int(x)
	case int64:
		return int(x)
	}
	// strconv fallback for a string status_code is best-effort; the
	// canonical paths always produce ints.
	if s, ok := v.(string); ok {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	}
	return 0
}
