package application

import (
	"strings"
	"testing"
)

var (
	configText = `
# config.yaml
# postgres DSN, ports number for rest, grpc message listener and grpc auth client

postgres:
  connection-string: "postgres://root:pass@127.0.0.1:5432/test-db"

ifaces:
  rest_port: 3000
  msg_listener_port: 4000
  auth_server_address: "auth:4000"

logger:
  level: debug
`

	cfgExpected = Config{
		Postgres: Postgres{
			DSN: "postgres://root:pass@127.0.0.1:5432/test-db",
		},
		IFaces: IFaces{
			RESTPort:    "3000",
			MSGPort:     "4000",
			AUTHAddress: "auth:4000",
		},
		Logger: Logger{
			Level: "debug",
		},
	}
)

func TestReadConfig(t *testing.T) {

	cfg := readConfigFile(strings.NewReader(configText))

	if cfgExpected != *cfg {
		t.Errorf("error reading config: expected %v, got %v", cfgExpected, *cfg)
	}
}
