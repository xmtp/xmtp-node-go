package openapi

import (
	_ "embed"
)

//go:embed message_api/v1/message_api.swagger.json
var JSON []byte
