package database

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// CreateUser creates a new user in the database
func CreateUser(email, name string) (*User, error) {
	token, err := generateUnsubscribeToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unsubscribe token: %w", err)
	}

	user := &User{
		Email:            email,
		Name:             name,
		Subscribed:       true,
		UnsubscribeToken: token,
	}

	result := DB.Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return user, nil
}

// GetAllSubscribedUsers returns all users who are subscribed
func GetAllSubscribedUsers() ([]User, error) {
	var users []User
	result := DB.Where("subscribed = ?", true).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get subscribed users: %w", result.Error)
	}
	return users, nil
}

// GetUserByEmail finds a user by their email address
func GetUserByEmail(email string) (*User, error) {
	var user User
	result := DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

// GetUserByToken finds a user by their unsubscribe token
func GetUserByToken(token string) (*User, error) {
	var user User
	result := DB.Where("unsubscribe_token = ?", token).First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

// UpdateUserSubscription updates a user's subscription status
func UpdateUserSubscription(userID uint, subscribed bool) error {
	result := DB.Model(&User{}).Where("id = ?", userID).Update("subscribed", subscribed)
	if result.Error != nil {
		return fmt.Errorf("failed to update user subscription: %w", result.Error)
	}
	return nil
}

// generateUnsubscribeToken generates a random token for unsubscribe links
func generateUnsubscribeToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
