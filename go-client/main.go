package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{
			"<ec2-ip>:9000",
			"<ec2-ip>:9001",
		},
		Auth: clickhouse.Auth{
			Database: "uk",
			Username: "default",
			Password: "",
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: map[string]interface{}{
			"send_logs_level": "trace",
		},
	})

	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	for {
		fmt.Println("Running analytics query...")
		ctx := clickhouse.Context(context.Background())

		rows, err := conn.Query(ctx, `
            SELECT
                town,
                district,
                count() AS c,
                round(avg(price)) AS price
            FROM uk.uk_price_paid
            WHERE date >= '2020-01-01'
            GROUP BY
                town,
                district
            HAVING c >= 100
            ORDER BY price DESC
            LIMIT 5
        `)

		if err != nil {
			log.Printf("Query failed: %v", err)
		} else {
			for rows.Next() {
				var town, district string
				var count uint64
				var price float64
				if err := rows.Scan(&town, &district, &count, &price); err != nil {
					log.Printf("Scan error: %v", err)
				}
				fmt.Printf("Town: %-15s District: %-15s Count: %-6d Price: %-8f\n", town, district, count, price)
			}
		}

		// Get the node that handled this query
		var host string
		if err := conn.QueryRow(ctx, `SELECT hostName()`).Scan(&host); err != nil {
			log.Printf("Failed to get host name: %v", err)
		} else {
			fmt.Printf("Query was served by node: %s\n", host)
		}

		fmt.Println("Sleeping 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}
