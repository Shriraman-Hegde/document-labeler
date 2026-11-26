package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client talks to a llama.cpp-compatible HTTP server or a test stub.
type Client struct {
	BaseURL    string
	AllowStub  bool
	HTTPClient *http.Client
}

// ConfigFromEnv builds a client; fails closed unless stub allowed.
func ConfigFromEnv() (*Client, error) {
	stub := os.Getenv("DIGITIZER_ALLOW_STUB") == "1"
	url := strings.TrimSpace(os.Getenv("DIGITIZER_LLAMA_URL"))
	if url == "" && !stub {
		return nil, fmt.Errorf("DIGITIZER_LLAMA_URL is required (or set DIGITIZER_ALLOW_STUB=1 for tests)")
	}
	return &Client{
		BaseURL:    strings.TrimRight(url, "/"),
		AllowStub:  stub,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}, nil
}

// AssistRequest is typing/schema help input.
type AssistRequest struct {
	OCRText  string
	Filename string
	WeakOCR  bool
}

// AssistResponse is model judgment.
type AssistResponse struct {
	DocType    string   `json:"docType"`
	Confidence float64  `json:"confidence"`
	Headings   []string `json:"headings"`
	Improved   string   `json:"improvedText,omitempty"`
}

// Assist classifies and optionally improves weak OCR text.
func (c *Client) Assist(req AssistRequest) (AssistResponse, error) {
	if c.AllowStub && c.BaseURL == "" {
		return stubAssist(req), nil
	}
	body := map[string]any{
		"messages": []map[string]string{
			{"role": "system", "content": "You label manufacturing scans. Reply JSON only: {\"docType\",\"confidence\",\"headings\",\"improvedText\"}. docType snake_case."},
			{"role": "user", "content": fmt.Sprintf("filename=%s weak=%v\nocr:\n%s", req.Filename, req.WeakOCR, truncate(req.OCRText, 4000))},
		},
		"temperature": 0.1,
	}
	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequest(http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return AssistResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	res, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		if c.AllowStub {
			return stubAssist(req), nil
		}
		return AssistResponse{}, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		if c.AllowStub {
			return stubAssist(req), nil
		}
		return AssistResponse{}, fmt.Errorf("llm http %s", res.Status)
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return AssistResponse{}, err
	}
	if len(parsed.Choices) == 0 {
		return AssistResponse{}, fmt.Errorf("empty llm response")
	}
	content := parsed.Choices[0].Message.Content
	content = extractJSON(content)
	var out AssistResponse
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		if c.AllowStub {
			return stubAssist(req), nil
		}
		return AssistResponse{}, fmt.Errorf("parse llm json: %w", err)
	}
	return out, nil
}

func stubAssist(req AssistRequest) AssistResponse {
	text := strings.ToLower(req.OCRText + " " + req.Filename)
	type guess struct {
		t    string
		keys []string
	}
	candidates := []guess{
		{"invoice", []string{"invoice"}},
		{"purchase_order", []string{"purchase order", "po number", "po#"}},
		{"packing_slip", []string{"packing", "packing slip"}},
		{"goods_receipt", []string{"goods receipt", "grn"}},
		{"inspection_report", []string{"inspection", "coa"}},
	}
	best := "unknown"
	conf := 0.35
	var headings []string
	for _, c := range candidates {
		for _, k := range c.keys {
			if strings.Contains(text, k) {
				best = c.t
				conf = 0.82
				headings = append(headings, k)
			}
		}
	}
	improved := req.OCRText
	if req.WeakOCR && improved == "" {
		improved = strings.ReplaceAll(filepathBase(req.Filename), "-", " ")
	}
	return AssistResponse{DocType: best, Confidence: conf, Headings: headings, Improved: improved}
}

func filepathBase(p string) string {
	i := strings.LastIndexAny(p, `/\`)
	if i >= 0 {
		return p[i+1:]
	}
	return p
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
