package judgepool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}

type JudgeRequest struct {
	Language          string     `json:"language"`
	SourceCode        string     `json:"source_code"`
	TimeLimitMs       int32      `json:"time_limit_ms"`
	MemoryLimitMb     int32      `json:"memory_limit_mb"`
	TestCases         []TestCase `json:"test_cases"`
	CheckerType       string     `json:"checker_type"`
	CustomCheckerCode string     `json:"custom_checker_code,omitempty"`
	RunAllTests       bool       `json:"run_all_tests"`
}

type TestResult struct {
	TestIndex int     `json:"test_index"`
	Verdict   string  `json:"verdict"`
	TimeMs    float64 `json:"time_ms"`
	MemoryKb  float64 `json:"memory_kb"`
	Stdout    string  `json:"stdout"`
	Stderr    string  `json:"stderr"`
}

type JudgeResponse struct {
	Verdict       string       `json:"verdict"`
	CompileOutput string       `json:"compile_output"`
	TestResults   []TestResult `json:"test_results"`
}

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second, // Max timeout for judging
		},
	}
}

func (c *Client) SubmitJudge(ctx context.Context, req JudgeRequest) (*JudgeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/judge", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("judge returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var judgeResp JudgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&judgeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &judgeResp, nil
}
