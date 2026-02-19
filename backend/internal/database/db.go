package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(databaseURL string) *pgxpool.Pool {
	maxRetries := 15
	for i := 0; i < maxRetries; i++ {
		pool, err := pgxpool.New(context.Background(), databaseURL)
		if err == nil {
			if err := pool.Ping(context.Background()); err == nil {
				fmt.Println("Connected to database")
				return pool
			} else {
				fmt.Printf("Ping error: %v\n", err)
			}
			pool.Close()
		} else {
			fmt.Printf("Connection error: %v\n", err)
		}
		fmt.Printf("Database not ready, retrying in 2s (%d/%d)...\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("Could not connect to database after retries")
	return nil
}
