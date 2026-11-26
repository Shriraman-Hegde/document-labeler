package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client talks to a llama.cpp-compatible OpenAI HTTP server or a test stub.
type Client struct {
	BaseURL    string
	Model      string
	AllowStub  bool
	HTTPClient *http.Client
}

// ConfigFromEnv builds a client; fails closed unless stub allowed.
func ConfigFromEnv() (*Client, error) {
	stub := os.Getenv("DIGITIZER_ALLOW_STUB") == "1"
	url := strings.TrimSpace(os.Getenv("DIGITIZER_LLAMA_URL"))
	model := strings.TrimSpace(os.Getenv("DIGITIZER_MODEL"))
	if model == "" {
		model = "qwen"
	}
	if url == "" && !stub {
		return nil, fmt.Errorf("DIGITIZER_LLAMA_URL is required (point at llama.cpp server, e.g. http://127.0.0.1:8080); or set DIGITIZER_ALLOW_STUB=1 for tests")
	}
	timeout := 120 * time.Second
	if v := os.Getenv("DIGITIZER_LLM_TIMEOUT_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeout = time.Duration(n) * time.Second
		}
	}
	return &Client{
		BaseURL:    strings.TrimRight(url, "/"),
		Model:      model,
		AllowStub:  stub,
		HTTPClient: &http.Client{Timeout: timeout},
	}, nil
}

// Ready checks the llama.cpp server is reachable (skipped for stub-only).
func (c *Client) Ready() error {
	if c.BaseURL == "" {
		if c.AllowStub {
			return nil
		}
		return fmt.Errorf("llm base URL empty")
	}
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/health", nil)
	if err != nil {
		return err
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		// Some llama.cpp builds expose /v1/models instead of /health.
		return c.readyModels()
	}
	defer res.Body.Close()
	if res.StatusCode >= 200 && res.StatusCode < 500 {
		return nil
	}
	return c.readyModels()
}

func (c *Client) readyModels() error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+"/v1/models", nil)
	if err != nil {
		return err
	}
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.AllowStub {
			return nil
		}
		return fmt.Errorf("llm unreachable at %s: %w (start llama.cpp or set DIGITIZER_ALLOW_STUB=1)", c.BaseURL, err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		if c.AllowStub {
			return nil
		}
		return fmt.Errorf("llm health check failed: %s", res.Status)
	}
	return nil
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
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": fmt.Sprintf("filename=%s weakOCR=%v\nocr:\n%s", req.Filename, req.WeakOCR, truncate(req.OCRText, 4000))},
		},
		"temperature": 0.1,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return AssistResponse{}, err
	}
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
		return AssistResponse{}, fmt.Errorf("llm request failed: %w", err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		if c.AllowStub {
			return stubAssist(req), nil
		}
		return AssistResponse{}, fmt.Errorf("llm http %s: %s", res.Status, truncate(string(raw), 200))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return AssistResponse{}, fmt.Errorf("llm response decode: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return AssistResponse{}, fmt.Errorf("empty llm response")
	}
	content := extractJSON(parsed.Choices[0].Message.Content)
	var out AssistResponse
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		if c.AllowStub {
			return stubAssist(req), nil
		}
		return AssistResponse{}, fmt.Errorf("parse llm json: %w (content=%s)", err, truncate(content, 120))
	}
	if out.DocType == "" {
		out.DocType = "unknown"
	}
	return out, nil
}

const systemPrompt = `You label manufacturing document scans.
Reply with JSON only, no markdown fences:
{"docType":"snake_case","confidence":0.0,"headings":["..."],"improvedText":"..."}
docType must be one of: invoice, purchase_order, packing_slip, goods_receipt, inspection_report, unknown.
If weakOCR is true, put cleaned readable text in improvedText; otherwise leave improvedText empty.`

func stubAssist(req AssistRequest) AssistResponse {
	text := strings.ToLower(req.OCRText + " " + req.Filename)
	type guess struct {
		t    string
		keys []string
	}
	candidates := []guess{
		{"invoice", []string{"invoice"}},
		{"purchase_order", []string{"purchase order", "po number", "po#"}},
		{"packing_slip", []string{"packing slip", "packing"}},
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
