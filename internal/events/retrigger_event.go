// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package trigger

import (
	"fmt"

	"github.com/twitchdev/twitch-cli/internal/util"
)

func RefireEvent(id string, p TriggerParamaters) (string, error) {
	res, err := util.GetEventByID(id)
	if err != nil {
		return "", err
	}

	p.Transport = res.Transport

	if p.ForwardAddress != "" {
		s, err := forwardEvent(ForwardParamters{
			ID:             id,
			Transport:      res.Transport,
			ForwardAddress: p.ForwardAddress,
			Secret:         p.Secret,
			JSON:           []byte(res.JSON),
			Event:          triggerTypeMap[res.Transport][res.Event],
		})
		if err != nil {
			return "", err
		}
		fmt.Printf("[%v] Endpoint received refired event.", s)
	}

	return res.JSON, nil
}
