package setup

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

var (
	postgresPool     *pgxpool.Pool
	postgresPoolOnce sync.Once
)

func ConnectDB() *pgxpool.Pool {
	postgresPoolOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		var err error

		databaseURL := os.Getenv("DATABASE_URL")
		postgresPool, err = pgxpool.New(ctx, databaseURL)
		if err != nil {
			log.Fatal("Unable to connect with the database\n", err)
		}

		if err = postgresPool.Ping(context.Background()); err != nil {
			log.Fatal("Unable to ping the database\n", err)
		}

		log.Printf("Connected to database! URL: %s\n", databaseURL)
	})

	return postgresPool
}
