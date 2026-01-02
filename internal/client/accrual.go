package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mrhyman/gophermart/api"
	"github.com/mrhyman/gophermart/internal/config"
	"github.com/mrhyman/gophermart/internal/logger"
	"github.com/mrhyman/gophermart/internal/model"
)

type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: normalizeBaseURL(baseURL),
		httpClient: &http.Client{
			Timeout: config.AccuralRequestTimeout,
		},
	}
}

func (c *AccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*api.AccrualResponse, error) {
	log := logger.FromContext(ctx)

	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	log.With("url", url).Warn()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.With("err", model.ErrAccrualRequestCreateFailed.Error()).Error()
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.With("err", model.ErrAccrualRequestSendFailed.Error()).Error()
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp api.AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&accrualResp); err != nil {
			log.With("err", model.ErrResponseDecode.Error()).Error()
			return nil, err
		}
		return &accrualResp, nil

	case http.StatusNoContent:
		return nil, model.ErrOrderNotRegistered

	case http.StatusTooManyRequests:
		log.With("err", model.ErrAccrualTooManyRequests.Error()).Warn()
		return nil, model.ErrAccrualTooManyRequests

	case http.StatusInternalServerError:
		return nil, model.ErrAccrualInternalError

	default:
		return nil, model.ErrWentWrong
	}
}

func normalizeBaseURL(baseURL string) string {
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}

	return baseURL
}
