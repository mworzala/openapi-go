// Code generated with openapi-go DO NOT EDIT.
package v1

import (
	"time"

	"github.com/google/uuid"
)

type GetMapWorldAbc struct {
	Page *int `json:"page"`
	Size *int `json:"size"`
}

type MapManualTriggerWebhookInner struct {
	A string `json:"a"`
}

type MapManualTriggerWebhook struct {
	Sender    string                        `json:"sender"`
	OrgId     string                        `json:"orgId"`
	MapId     uuid.UUID                     `json:"mapId"`
	Timestamp time.Time                     `json:"timestamp"`
	Inner     *MapManualTriggerWebhookInner `json:"inner"`
}

type GetMapWorldResponse struct {
	Polar []byte `json:"Polar"`
	Anvil []byte `json:"Anvil"`
}
