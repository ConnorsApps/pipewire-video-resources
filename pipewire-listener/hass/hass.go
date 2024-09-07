package hass

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

type Client struct {
	resty *resty.Client
}

func New(url, token string) *Client {
	return &Client{
		resty: resty.New().
			SetBaseURL(url).
			SetRetryCount(3).
			SetRetryWaitTime(2 * time.Second).
			SetAuthToken(token),
	}
}

func (c *Client) jsonRequest(domain, method string, request interface{}) error {
	resp, err := c.resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		Post(fmt.Sprintf("api/services/%s/%s", domain, method))

	if err != nil || resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		log.Error().Err(err).
			Interface("request", request).
			Int("statusCode", resp.StatusCode()).
			Interface("reponse", resp).
			Str("domain", domain).
			Msg("bad status code for home assistant request")

		return fmt.Errorf("hass request: %v", err)
	}
	return nil
}

func (c *Client) entityState(entityId string) ([]byte, error) {
	resp, err := c.resty.R().
		Get("api/states/" + entityId)

	if err != nil || resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		log.Error().Err(err).
			Int("statusCode", resp.StatusCode()).
			Any("reponse", resp).
			Msg("bad status code for get state request")

		return nil, fmt.Errorf("hass request: %v", err)
	}

	return resp.Body(), nil
}

func (c *Client) TurnSwitchOn(deviceId string) error {
	log.Debug().Str("device_id", deviceId).Msg("Turning switch on")
	return c.jsonRequest("switch", "turn_on", map[string]string{"device_id": deviceId})
}

func (c *Client) TurnSwitchOff(deviceId string) error {
	log.Debug().Str("device_id", deviceId).Msg("Turning switch off")
	return c.jsonRequest("switch", "turn_off", map[string]string{"device_id": deviceId})
}

type SpeakerState string

const (
	SpeakerOn  SpeakerState = "on"
	SpeakerOff SpeakerState = "off"
)

type SwitchStatus struct {
	EntityID   string       `json:"entity_id"`
	State      SpeakerState `json:"state"`
	Attributes struct {
		FriendlyName string `json:"friendly_name"`
	} `json:"attributes"`
	LastChanged string `json:"last_changed"`
	LastUpdated string `json:"last_updated"`
	Context     struct {
		ID       string `json:"id"`
		ParentID string `json:"parent_id"`
		UserID   string `json:"user_id"`
	} `json:"context"`
}

func (c *Client) SwitchState(entityId string) (*SwitchStatus, error) {
	var status SwitchStatus
	state, err := c.entityState(entityId)
	if err != nil {
		return &status, err
	}

	if err := json.Unmarshal(state, &status); err != nil {
		return &status, err
	}

	return &status, nil
}
