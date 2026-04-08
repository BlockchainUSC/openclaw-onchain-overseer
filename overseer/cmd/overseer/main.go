package main

import (
	"fmt"

	_ "github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv"
	_ "github.com/tmc/langchaingo/llms"
)

func main() {
	fmt.Println("openclaw oversight agent")
}
