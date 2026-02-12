package types

// JWTConfig configuración de JWT
type JWTConfig struct {
	Secret            string
	Expiration        int
	RefreshSecret     string
	RefreshExpiration int
}
