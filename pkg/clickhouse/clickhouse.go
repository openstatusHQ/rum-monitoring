package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func NewClient() (driver.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"127.0.0.1:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
			// for  dev
			// TLS: &tls.Config{
			// 	InsecureSkipVerify: true,
			// },
			Settings: clickhouse.Settings{
				"max_execution_time": 60,
			},
			DialTimeout: time.Second * 30,
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			Debug:                true,
			BlockBufferSize:      10,
			MaxCompressionBuffer: 10240,
			ClientInfo: clickhouse.ClientInfo{ // optional, please see Client info section in the README.md
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "openstatus", Version: "0.1"},
				},
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}
