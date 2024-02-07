// Code generated with openapi-go DO NOT EDIT.
package v1

import "time"

type MapManualTriggerWebhook struct {
	Sender    string    `json:"sender"`
	OrgId     string    `json:"orgId"`
	MapId     string    `json:"mapId"`
	Timestamp time.Time `json:"timestamp"`
}

type GetMapWorldResponse struct {
	Polar []byte `json:"Polar"`
	Anvil []byte `json:"Anvil"`
}
