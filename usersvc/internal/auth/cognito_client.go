package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws_e2e_test/usersvc/internal/model"
)

// CognitoClient handles authentication with AWS Cognito
type CognitoClient struct {
	client           *cognitoidentityprovider.Client
	userPoolID       string
	userPoolClientID string
}

// NewCognitoClient creates a new Cognito client
func NewCognitoClient(region, userPoolID, userPoolClientID string) (*CognitoClient, error) {
	log.Printf("Initializing Cognito client with region: %s, user pool ID: %s, client ID: %s",
		region, userPoolID, userPoolClientID)

	// Validate inputs
	if region == "" {
		return nil, fmt.Errorf("region cannot be empty")
	}
	if userPoolID == "" {
		return nil, fmt.Errorf("user pool ID cannot be empty")
	}
	if userPoolClientID == "" {
		return nil, fmt.Errorf("user pool client ID cannot be empty")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		log.Printf("Failed to load AWS config: %v", err)
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Cognito client
	client := cognitoidentityprovider.NewFromConfig(cfg)

	return &CognitoClient{
		client:           client,
		userPoolID:       userPoolID,
		userPoolClientID: userPoolClientID,
	}, nil
}

// SignUp registers a new user with Cognito
func (c *CognitoClient) SignUp(email, password, firstName, lastName string) error {
	log.Printf("Signing up user with email: %s", email)

	// Create the sign-up request
	input := &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(c.userPoolClientID),
		Username: aws.String(email),
		Password: aws.String(password),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(email),
			},
			{
				Name:  aws.String("given_name"),
				Value: aws.String(firstName),
			},
			{
				Name:  aws.String("family_name"),
				Value: aws.String(lastName),
			},
		},
	}

	// Call Cognito to sign up the user
	_, err := c.client.SignUp(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to sign up user: %v", err)
		return fmt.Errorf("failed to sign up user: %w", err)
	}

	log.Printf("Successfully signed up user with email: %s", email)
	return nil
}

// ConfirmSignUp confirms a user's registration with the confirmation code
func (c *CognitoClient) ConfirmSignUp(email, confirmationCode string) error {
	log.Printf("Confirming sign up for user with email: %s", email)

	// Create the confirm sign-up request
	input := &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(c.userPoolClientID),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(confirmationCode),
	}

	// Call Cognito to confirm the user
	_, err := c.client.ConfirmSignUp(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to confirm sign up: %v", err)
		return fmt.Errorf("failed to confirm sign up: %w", err)
	}

	log.Printf("Successfully confirmed sign up for user with email: %s", email)
	return nil
}

// ResendConfirmationCode resends the confirmation code to the user
func (c *CognitoClient) ResendConfirmationCode(email string) error {
	log.Printf("Resending confirmation code for user with email: %s", email)

	// Create the resend confirmation code request
	input := &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId: aws.String(c.userPoolClientID),
		Username: aws.String(email),
	}

	// Call Cognito to resend the confirmation code
	_, err := c.client.ResendConfirmationCode(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to resend confirmation code: %v", err)
		return fmt.Errorf("failed to resend confirmation code: %w", err)
	}

	log.Printf("Successfully resent confirmation code for user with email: %s", email)
	return nil
}

// Login authenticates a user and returns the authentication tokens
func (c *CognitoClient) Login(email, password string) (*model.AuthResponse, error) {
	log.Printf("Logging in user with email: %s", email)

	// Create the authentication request
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(c.userPoolClientID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	// Call Cognito to authenticate the user
	result, err := c.client.InitiateAuth(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to authenticate user: %v", err)
		return nil, fmt.Errorf("failed to authenticate user: %w", err)
	}

	// Extract the authentication tokens
	authResult := result.AuthenticationResult
	if authResult == nil {
		log.Printf("Authentication result is nil")
		return nil, fmt.Errorf("authentication result is nil")
	}

	// Create the authentication response
	response := &model.AuthResponse{
		AccessToken:  *authResult.AccessToken,
		IdToken:      *authResult.IdToken,
		RefreshToken: *authResult.RefreshToken,
		ExpiresIn:    int(authResult.ExpiresIn),
		TokenType:    *authResult.TokenType,
	}

	log.Printf("Successfully authenticated user with email: %s", email)
	return response, nil
}

// RefreshToken refreshes the authentication tokens
func (c *CognitoClient) RefreshToken(refreshToken string) (*model.AuthResponse, error) {
	log.Printf("Refreshing authentication tokens")

	// Create the refresh token request
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshToken,
		ClientId: aws.String(c.userPoolClientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	}

	// Call Cognito to refresh the tokens
	result, err := c.client.InitiateAuth(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to refresh tokens: %v", err)
		return nil, fmt.Errorf("failed to refresh tokens: %w", err)
	}

	// Extract the authentication tokens
	authResult := result.AuthenticationResult
	if authResult == nil {
		log.Printf("Authentication result is nil")
		return nil, fmt.Errorf("authentication result is nil")
	}

	// Create the authentication response
	response := &model.AuthResponse{
		AccessToken: *authResult.AccessToken,
		IdToken:     *authResult.IdToken,
		ExpiresIn:   int(authResult.ExpiresIn),
		TokenType:   *authResult.TokenType,
	}

	// The refresh token might not be returned if it hasn't changed
	if authResult.RefreshToken != nil {
		response.RefreshToken = *authResult.RefreshToken
	} else {
		response.RefreshToken = refreshToken
	}

	log.Printf("Successfully refreshed authentication tokens")
	return response, nil
}

// ForgotPassword initiates the forgot password flow
func (c *CognitoClient) ForgotPassword(email string) error {
	log.Printf("Initiating forgot password flow for user with email: %s", email)

	// Create the forgot password request
	input := &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: aws.String(c.userPoolClientID),
		Username: aws.String(email),
	}

	// Call Cognito to initiate the forgot password flow
	_, err := c.client.ForgotPassword(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to initiate forgot password flow: %v", err)
		return fmt.Errorf("failed to initiate forgot password flow: %w", err)
	}

	log.Printf("Successfully initiated forgot password flow for user with email: %s", email)
	return nil
}

// ConfirmForgotPassword completes the forgot password flow
func (c *CognitoClient) ConfirmForgotPassword(email, confirmationCode, newPassword string) error {
	log.Printf("Confirming forgot password for user with email: %s", email)

	// Create the confirm forgot password request
	input := &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(c.userPoolClientID),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(confirmationCode),
		Password:         aws.String(newPassword),
	}

	// Call Cognito to confirm the forgot password
	_, err := c.client.ConfirmForgotPassword(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to confirm forgot password: %v", err)
		return fmt.Errorf("failed to confirm forgot password: %w", err)
	}

	log.Printf("Successfully confirmed forgot password for user with email: %s", email)
	return nil
}

// ChangePassword changes the password for an authenticated user
func (c *CognitoClient) ChangePassword(accessToken, oldPassword, newPassword string) error {
	log.Printf("Changing password for authenticated user")

	// Create the change password request
	input := &cognitoidentityprovider.ChangePasswordInput{
		AccessToken:      aws.String(accessToken),
		PreviousPassword: aws.String(oldPassword),
		ProposedPassword: aws.String(newPassword),
	}

	// Call Cognito to change the password
	_, err := c.client.ChangePassword(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to change password: %v", err)
		return fmt.Errorf("failed to change password: %w", err)
	}

	log.Printf("Successfully changed password for authenticated user")
	return nil
}

// GetUser gets the user attributes for an authenticated user
func (c *CognitoClient) GetUser(accessToken string) (map[string]string, error) {
	log.Printf("Getting user attributes for authenticated user")

	// Create the get user request
	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(accessToken),
	}

	// Call Cognito to get the user
	result, err := c.client.GetUser(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to get user: %v", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Extract the user attributes
	attributes := make(map[string]string)
	for _, attr := range result.UserAttributes {
		attributes[*attr.Name] = *attr.Value
	}

	log.Printf("Successfully got user attributes for authenticated user")
	return attributes, nil
}

// UpdateUserAttributes updates the user attributes for an authenticated user
func (c *CognitoClient) UpdateUserAttributes(accessToken string, attributes map[string]string) error {
	log.Printf("Updating user attributes for authenticated user")

	// Convert the attributes to the Cognito format
	userAttributes := make([]types.AttributeType, 0, len(attributes))
	for name, value := range attributes {
		userAttributes = append(userAttributes, types.AttributeType{
			Name:  aws.String(name),
			Value: aws.String(value),
		})
	}

	// Create the update user attributes request
	input := &cognitoidentityprovider.UpdateUserAttributesInput{
		AccessToken:    aws.String(accessToken),
		UserAttributes: userAttributes,
	}

	// Call Cognito to update the user attributes
	_, err := c.client.UpdateUserAttributes(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to update user attributes: %v", err)
		return fmt.Errorf("failed to update user attributes: %w", err)
	}

	log.Printf("Successfully updated user attributes for authenticated user")
	return nil
}

// DeleteUser deletes the authenticated user
func (c *CognitoClient) DeleteUser(accessToken string) error {
	log.Printf("Deleting authenticated user")

	// Create the delete user request
	input := &cognitoidentityprovider.DeleteUserInput{
		AccessToken: aws.String(accessToken),
	}

	// Call Cognito to delete the user
	_, err := c.client.DeleteUser(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("Successfully deleted authenticated user")
	return nil
}

// AdminDeleteUser deletes a user as an administrator
func (c *CognitoClient) AdminDeleteUser(email string) error {
	log.Printf("Deleting user with email: %s as administrator", email)

	// Create the admin delete user request
	input := &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	}

	// Call Cognito to delete the user
	_, err := c.client.AdminDeleteUser(context.TODO(), input)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	log.Printf("Successfully deleted user with email: %s", email)
	return nil
}
