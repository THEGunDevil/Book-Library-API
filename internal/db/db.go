package db

import (
	"context"
	"log"

	"github.com/THEGunDevil/GoForBackend/internal/config"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	DB  *pgxpool.Pool
	Ctx = context.Background()
	Q   *gen.Queries
)

func Connect(cfg config.Config) {
	pool, err := pgxpool.New(Ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("❌ Unable to connect to database: %v", err)
	}

	if err := pool.Ping(Ctx); err != nil {
		log.Fatalf("❌ Could not ping database: %v", err)
	}

	DB = pool
	Q = gen.New(pool)
	log.Println("✅ Connected to Postgres successfully")
}

func Close() {
	if DB != nil {
		DB.Close()
		log.Println("🛑 Database connection closed")
	}
}
