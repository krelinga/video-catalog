package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/krelinga/video-catalog/internal"
)

type Server struct {
	Config *internal.Config
	Pool   *pgxpool.Pool
}
