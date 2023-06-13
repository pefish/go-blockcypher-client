package blockcypher_client

import (
	go_logger "github.com/pefish/go-logger"
	"testing"
	"time"
)

func TestBlockcypherClient_ListUnspent(t *testing.T) {
	type fields struct {
		timeout time.Duration
		logger  go_logger.InterfaceLogger
		baseUrl string
		key     string
	}
	type args struct {
		address string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []ListUnspentResult
		wantErr bool
	}{
		{
			args: struct{ address string }{address: "1DEP8i3QJCsomS4BSMY2RpU1upv62aGvhD"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBlockcypherClient(go_logger.Logger, 5*time.Second, "")
			got, err := bc.ListUnspent(tt.args.address)
			go_logger.Logger.Info(got, err)
		})
	}
}
