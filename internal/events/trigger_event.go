// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package trigger

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/twitchdev/twitch-cli/internal/util"
)

// TriggerParamaters defines the parameters used to emit an event.
type TriggerParamaters struct {
	Event          string
	Transport      string
	IsAnonymous    bool
	FromUser       string
	ToUser         string
	GiftUser       string
	ForwardAddress string
	Secret         string
	Verbose        bool
	Count          int
}

type TriggerResponse struct {
	ID        string
	JSON      []byte
	FromUser  string
	ToUser    string
	Timestamp string
}

// Fire emits an event using the TriggerParamaters defined above.
func Fire(p TriggerParamaters) (string, error) {
	if len(triggerTypeMap[p.Transport]) == 0 {
		return "", errors.New("Invalid transport")
	}

	if triggerTypeMap[p.Transport][p.Event] == "" {
		return "", errors.New("Event unsupported for given transport")
	}

	event := triggerTypeMap[p.Transport][p.Event]

	var resp TriggerResponse
	var err error

	switch p.Event {
	// sub events
	case "subscribe", "unsubscribe", "gift":
		isGift := false
		if p.Event == "gift" {
			isGift = true
		}

		resp, err = GenerateSubBody(SubscribeParams{
			Transport:       p.Transport,
			Type:            event,
			IsGift:          isGift,
			IsAnonymousGift: p.IsAnonymous,
			ToUser:          p.ToUser,
			FromUser:        p.FromUser,
		})

	// bits events
	case "cheer":
		resp, err = GenerateCheerBody(CheerParams{
			Transport:   p.Transport,
			Type:        event,
			FromUser:    p.FromUser,
			ToUser:      p.ToUser,
			IsAnonymous: p.IsAnonymous,
		})

	case "transaction":
		resp, err = GenerateTransactionBody(TransactionParams{
			Transport: p.Transport,
			Type:      event,
			FromUser:  p.FromUser,
			ToUser:    p.ToUser,
		})
	default:
		return "", nil
	}
	if err != nil {
		return "", err
	}

	err = util.InsertIntoDB(util.EventCacheParameters{
		ID:        resp.ID,
		Event:     p.Event,
		JSON:      string(resp.JSON),
		FromUser:  resp.FromUser,
		ToUser:    resp.ToUser,
		Transport: p.Transport,
		Timestamp: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}

	if p.ForwardAddress != "" {
		statusCode, err := forwardEvent(ForwardParamters{
			ID:             resp.ID,
			Transport:      p.Transport,
			JSON:           resp.JSON,
			Secret:         p.Secret,
			ForwardAddress: p.ForwardAddress,
			Event:          event,
		})

		if err != nil {
			return "", err
		}

		println(fmt.Sprintf(`[%v] Request Sent`, statusCode))
	}

	return string(resp.JSON), nil
}

func ValidTriggers() []string {
	names := []string{}

	for name, enabled := range triggerSupported {
		if enabled == true {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}

func ValidTransports() []string {
	names := []string{}

	for name, enabled := range transportSupported {
		if enabled == true {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	return names
}
