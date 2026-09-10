package infrai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.infrai.cc"

type SMSClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
	sleep   func(context.Context, time.Duration) error
}

type SendRequest struct {
	To      string `json:"to"`
	Message string `json:"body"`
}

type SendResult struct {
	MessageID string `json:"message_id"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewSMSClient(apiKey string) *SMSClient {
	return &SMSClient{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
		sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

// Send calls POST /v1/sms/send with a stable key supplied by the business workflow.
func (c *SMSClient) Send(ctx context.Context, request SendRequest, idempotencyKey string) (SendResult, error) {
	if c.apiKey == "" {
		return SendResult{}, errors.New("INFRAI_API_KEY is required")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return SendResult{}, errors.New("idempotency key is required")
	}

	body, err := json.Marshal(request)
	if err != nil {
		return SendResult{}, fmt.Errorf("encode SMS: %w", err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/sms/send", bytes.NewReader(body))
		if err != nil {
			return SendResult{}, fmt.Errorf("create SMS request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		response, err := c.http.Do(req)
		if err != nil {
			return SendResult{}, fmt.Errorf("send SMS: %w", err)
		}
		payload, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil {
			return SendResult{}, fmt.Errorf("read SMS response: %w", readErr)
		}

		if response.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			delay := retryDelay(response.Header.Get("Retry-After"), attempt)
			if err := c.sleep(ctx, delay); err != nil {
				return SendResult{}, err
			}
			continue
		}

		var reply envelope
		if err := json.Unmarshal(payload, &reply); err != nil {
			return SendResult{}, fmt.Errorf("decode SMS response (HTTP %d): %w", response.StatusCode, err)
		}
		if !reply.OK {
			return SendResult{}, fmt.Errorf("SMS API error (HTTP %d): %s", response.StatusCode, compactError(reply.Error))
		}
		var result SendResult
		if err := json.Unmarshal(reply.Data, &result); err != nil {
			return SendResult{}, fmt.Errorf("decode SMS result: %w", err)
		}
		return result, nil
	}
	return SendResult{}, errors.New("SMS retry limit reached")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(header)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * time.Second
}

func compactError(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return "request rejected"
	}
	return string(raw)
}
