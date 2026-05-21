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
		cancel()

		var err error

		databaseURL := os.Getenv("DATABASE_URL")
		log.Println(databaseURL)
		postgresPool, err = pgxpool.New(ctx, databaseURL)
		if err != nil {
			log.Fatal("Unable to connect with the database\n", err)
		}

		if err = postgresPool.Ping(context.Background()); err != nil {
			log.Fatal("Unable to ping the database\n", err)
		}

		log.Println("Connected to database!")
	})

	return postgresPool
}
