package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
)

var (
	listen = flag.String("listen", ":8080", "Address to listen on")
	// List of common text-based MIME types
	textContentTypes = map[string]bool{
		"text/plain":             true,
		"text/html":              true,
		"text/css":               true,
		"text/xml":               true,
		"text/csv":               true,
		"text/javascript":        true,
		"text/markdown":          true,
		"application/json":       true,
		"application/xml":        true,
		"application/ld+json":    true,
		"application/javascript": true, // RFC 4329, though text/javascript is more common
		"application/xhtml+xml":  true,
		"application/atom+xml":   true,
		"application/rss+xml":    true,
		"image/svg+xml":          true, // SVG is XML-based and human-readable
	}
)

// isClientOrServerError checks if the status code is a 4xx or 5xx error.
func isClientOrServerError(statusCode int) bool {
	return statusCode >= 400 && statusCode <= 599
}

// isTextContentType checks if the given Content-Type string indicates a textual format.
func isTextContentType(contentTypeHeader string) bool {
	// Normalize: to lowercase and take only the main type/subtype, ignore parameters like charset
	mediaType := strings.ToLower(strings.Split(contentTypeHeader, ";")[0])

	// Check if the media type is in the predefined list of text content types
	if textContentTypes[mediaType] {
		return true
	}

	// Catch-all for other "text/*" types
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}

	return false
}

func modifyResponse(resp *http.Response) error {
	log.Printf("Response from upstream for %s: Status %s\n", resp.Request.URL, resp.Status)

	contentType := resp.Header.Get("Content-Type")
	if isClientOrServerError(resp.StatusCode) || isTextContentType(contentType) {
		statusStr := "SUCCESSFUL"
		if resp.StatusCode >= 400 {
			statusStr = "FAILED"
		}
		log.Printf("Upstream request to %s %s with status: %s. Dumping request and response.\n", resp.Request.URL, statusStr, resp.Status)
		dump, err := httputil.DumpRequestOut(resp.Request, true) // true to include body
		if err != nil {
			log.Printf("Error dumping %s request for %s: %v\n", statusStr, resp.Request.URL, err)
			// Fallback logging if dump fails
			fmt.Fprintf(os.Stdout, "---- BEGIN %s UPSTREAM REQUEST (DUMP FAILED: %v) ----\nURL: %s\nMethod: %s\nHeaders: %v\n---- END %s UPSTREAM REQUEST ----\n", statusStr, err, resp.Request.URL, resp.Request.Method, resp.Request.Header, statusStr)
			return nil // Allow the original error response to proceed if dump fails
		}
		fmt.Fprintf(os.Stdout, "---- BEGIN %s UPSTREAM REQUEST DUMP (URL: %s) ----\n%s\n---- END %s UPSTREAM REQUEST DUMP ----\n", statusStr, resp.Request.URL, string(dump), statusStr)
		// DumpResponse reads and replaces resp.Body with a new ReadCloser.
		dump, err = httputil.DumpResponse(resp, true) // true to include body
		if err != nil {
			log.Printf("Error dumping %s response for %s: %v\n", statusStr, resp.Request.URL, err)
			// Fallback logging if dump fails
			fmt.Fprintf(os.Stdout, "---- BEGIN %s UPSTREAM RESPONSE (DUMP FAILED: %v) ----\nURL: %s\nStatus: %s\n---- END %s UPSTREAM RESPONSE ----\n", statusStr, err, resp.Request.URL, resp.Status, statusStr)
			return nil // Allow the original error response to proceed if dump fails
		}
		fmt.Fprintf(os.Stdout, "---- BEGIN %s UPSTREAM RESPONSE DUMP (URL: %s) ----\n%s\n---- END %s UPSTREAM RESPONSE DUMP ----\n", statusStr, resp.Request.URL, string(dump), statusStr)
		// The body is already replaced by DumpResponse, so ReverseProxy can send it.
	} else {
		// For other statuses (1xx informational, 3xx redirection), just log.
		// Redirections are typically handled by the client (browser) based on the headers.
		// ReverseProxy will pass these responses through.
		log.Printf("Response from %s with status %s (not dumping content for this type).\n", resp.Request.URL, resp.Status)
	}

	return nil // No error modifying response, let ReverseProxy handle it
}

func main() {
	director := func(req *http.Request) {
		// httputil.ReverseProxy will automatically set X-Forwarded-For, X-Forwarded-Host, X-Forwarded-Proto.
		// It uses req.URL.Host as the destination.
		// We need to ensure req.URL.Scheme and req.URL.Host are correctly set.
		// if req.URL.Scheme == "" {
		// 	req.URL.Scheme = "http" // Assume http if not specified
		// }
		req.URL.Scheme = "https" // Force HTTPS
		// The Host header is already set by the client to the target server.
		// ReverseProxy uses req.URL.Host for dialing.
		// Ensure req.URL.Host is set to the target. It usually is from r.RequestURI.
		log.Printf("Proxying %s request for: %s\n", req.Method, req.URL.String())
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error for %s: %v\n", r.URL, err)
		// This error means the upstream was unreachable (e.g., connection refused, DNS lookup failure).
		// There is no HTTP response from upstream to dump.
		fmt.Fprintf(os.Stdout, "---- FAILED UPSTREAM ATTEMPT ----\nURL: %s\nError: %v\n---- END FAILED UPSTREAM ATTEMPT ----\n", r.URL, err)
		http.Error(w, fmt.Sprintf("Upstream server error: %v", err), http.StatusBadGateway)
	}

	reverseProxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
		ErrorHandler:   errorHandler,
		Transport:      http.DefaultTransport, // You can customize transport (e.g., for timeouts)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			http.Error(w, "CONNECT method not supported", http.StatusMethodNotAllowed)
			return
		} else {
			reverseProxy.ServeHTTP(w, r)
		}
	})

	log.Printf("Starting HTTP proxy on %s\n", *listen)
	server := &http.Server{
		Addr:         *listen,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start proxy server: %v\n", err)
	}
}
