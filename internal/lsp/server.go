package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
}

// StartServer starts the stdio LSP server loop
func StartServer() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "Content-Length:") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		contentLength, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			continue
		}

		// Read trailing empty line
		_, _ = reader.ReadString('\n')

		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return err
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			continue
		}

		handleRequest(req)
	}
}

func handleRequest(req Request) {
	switch req.Method {
	case "initialize":
		res := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"capabilities": map[string]interface{}{
					"textDocumentSync": 1,
				},
			},
		}
		sendResponse(res)
	}
}

func sendResponse(res Response) {
	data, _ := json.Marshal(res)
	msg := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
	fmt.Print(msg)
}
