package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a set of JSON Web Keys
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWTValidatorConfig holds configuration for JWT validation
type JWTValidatorConfig struct {
	JWKSURL string
	Issuer  string
}

// JWTValidator handles JWT token validation
type JWTValidator struct {
	jwksURL string
	issuer  string
	keys    map[string]*rsa.PublicKey
}

// NewJWTValidator creates a new JWT validator with the provided configuration
func NewJWTValidator(config JWTValidatorConfig) *JWTValidator {
	return &JWTValidator{
		jwksURL: config.JWKSURL,
		issuer:  config.Issuer,
		keys:    make(map[string]*rsa.PublicKey),
	}
}

// NewCognitoJWTValidator creates a new JWT validator configured for AWS Cognito
func NewCognitoJWTValidator(region, userPoolID string) *JWTValidator {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", region, userPoolID)
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolID)

	return &JWTValidator{
		jwksURL: jwksURL,
		issuer:  issuer,
		keys:    make(map[string]*rsa.PublicKey),
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *JWTValidator) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Parse the token without verification first to get the kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get the key ID from the token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("token missing kid header")
	}

	// Get the public key for this kid
	publicKey, err := v.getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate the token
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	// Validate token type (should be "access" for access tokens)
	tokenUse, ok := claims["token_use"].(string)
	if !ok || tokenUse != "access" {
		return nil, fmt.Errorf("invalid token use: expected 'access', got '%s'", tokenUse)
	}

	// Validate issuer if provided
	if v.issuer != "" {
		iss, ok := claims["iss"].(string)
		if !ok || iss != v.issuer {
			return nil, fmt.Errorf("invalid issuer: expected '%s', got '%s'", v.issuer, iss)
		}
	}

	// Validate expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("token missing exp claim")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// getPublicKey retrieves the public key for the given kid
func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check if we already have this key cached
	if key, exists := v.keys[kid]; exists {
		return key, nil
	}

	// Fetch the JWKS
	jwks, err := v.fetchJWKS()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Find the key with the matching kid
	var jwk *JWK
	for _, key := range jwks.Keys {
		if key.Kid == kid {
			jwk = &key
			break
		}
	}

	if jwk == nil {
		return nil, fmt.Errorf("key with kid '%s' not found", kid)
	}

	// Convert JWK to RSA public key
	publicKey, err := v.jwkToRSAPublicKey(jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
	}

	// Cache the key
	v.keys[kid] = publicKey

	return publicKey, nil
}

// fetchJWKS fetches the JSON Web Key Set from the JWKS URL
func (v *JWTValidator) fetchJWKS() (*JWKSet, error) {
	resp, err := http.Get(v.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func (v *JWTValidator) jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create the RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// GetUserEmailFromClaims extracts the user email from JWT claims
func GetUserEmailFromClaims(claims jwt.MapClaims) (string, bool) {
	email, ok := claims["email"].(string)
	return email, ok
}

// GetUsernameFromClaims extracts the username from JWT claims
func GetUsernameFromClaims(claims jwt.MapClaims) (string, bool) {
	username, ok := claims["username"].(string)
	return username, ok
}

// GetUserSubFromClaims extracts the user subject (unique ID) from JWT claims
func GetUserSubFromClaims(claims jwt.MapClaims) (string, bool) {
	sub, ok := claims["sub"].(string)
	return sub, ok
}
