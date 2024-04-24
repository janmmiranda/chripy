module github.com/janmmiranda/chripy

go 1.22.2

require golang.org/x/crypto v0.22.0 // indirect

require github.com/golang-jwt/jwt/v5 v5.2.1 // indirect

require (
	github.com/joho/godotenv v1.5.1
)

require github.com/janmmiranda/chripy/internal/auth v0.0.0

replace github.com/janmmiranda/chripy/internal/auth => ./internal/auth
