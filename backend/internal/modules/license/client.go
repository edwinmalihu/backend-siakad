package license

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GeneratorClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewGeneratorClient(baseURL, apiKey string) *GeneratorClient {
	return &GeneratorClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *GeneratorClient) Validate(ctx context.Context, licenseKey, deviceFingerprint string) (*GeneratorValidateResponse, error) {
	body := map[string]string{
		"license_key":        licenseKey,
		"device_fingerprint": deviceFingerprint,
	}
	var resp GeneratorValidateResponse
	if err := c.do(ctx, "POST", "/api/validate", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *GeneratorClient) Activate(ctx context.Context, licenseKey, deviceFingerprint, clientName string) (*GeneratorActivateResponse, error) {
	body := map[string]string{
		"license_key":        licenseKey,
		"device_fingerprint": deviceFingerprint,
		"client_name":        clientName,
	}
	var resp GeneratorActivateResponse
	if err := c.do(ctx, "POST", "/api/activate", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *GeneratorClient) Trial(ctx context.Context, deviceFingerprint, clientName string, trialCount int) (*GeneratorTrialResponse, error) {
	body := map[string]interface{}{
		"device_fingerprint": deviceFingerprint,
		"client_name":        clientName,
		"trial_count":        trialCount,
	}
	var resp GeneratorTrialResponse
	if err := c.do(ctx, "POST", "/api/trial", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *GeneratorClient) Status(ctx context.Context, licenseKey string) (*GeneratorStatusResponse, error) {
	var resp GeneratorStatusResponse
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/status?key=%s", licenseKey), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *GeneratorClient) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("generator API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		// Generator API wraps responses in {"success":true,"data":{...}}
		var envelope struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
			Message string          `json:"message"`
		}
		if err := json.Unmarshal(respBody, &envelope); err != nil {
			return fmt.Errorf("unmarshal envelope: %w", err)
		}
		if !envelope.Success {
			msg := envelope.Message
			if msg == "" {
				msg = "generator returned success=false"
			}
			return fmt.Errorf("%s", msg)
		}
		if len(envelope.Data) > 0 {
			if err := json.Unmarshal(envelope.Data, result); err != nil {
				return fmt.Errorf("unmarshal data: %w", err)
			}
		}
	}

	return nil
}
