package main

type apiConfig struct {
	fileServerHits int
	DB             *DB
	JWTSecret      string
	PolkaKey       string
}
