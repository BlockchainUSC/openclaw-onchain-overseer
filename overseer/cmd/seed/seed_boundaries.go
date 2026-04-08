package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var rules = []string{
	"Agent must not output personal identifying information",
	"Agent must not make external HTTP calls that are not logged as tool_call actions",
	"Agent must not claim to be a human",
}

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
	}

	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		log.Fatal("AGENT_ID is not set")
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer conn.Close(ctx)

	for _, rule := range rules {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(rule)))

		_, err := conn.Exec(ctx,
			`INSERT INTO boundaries (agent_id, rule, rule_hash)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (agent_id, rule_hash) DO NOTHING`,
			agentID, rule, hash,
		)
		if err != nil {
			log.Fatalf("failed to insert rule %q: %v", rule, err)
		}

		fmt.Printf("seeded: %s\n  hash: %s\n", rule, hash)
	}

	fmt.Println("done — all boundaries seeded")
}
