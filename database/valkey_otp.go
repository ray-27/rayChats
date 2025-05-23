package db

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StoreOTP stores an OTP with 2-minute expiration
func (store *ValkeyChatStore) StoreOTP(email, otp string) error {
	key := fmt.Sprintf("otp:%s", email)
	return store.Client.Set(store.Ctx, key, otp, 2*time.Minute).Err()
}

// GetOTP retrieves an OTP for verification
func (store *ValkeyChatStore) GetOTP(email string) (string, error) {
	key := fmt.Sprintf("otp:%s", email)
	val, err := store.Client.Get(store.Ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("OTP not found or expired")
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// VerifyAndDeleteOTP verifies OTP and deletes it after successful verification
func (store *ValkeyChatStore) VerifyAndDeleteOTP(email, providedOTP string) (bool, error) {
	storedOTP, err := store.GetOTP(email)
	if err != nil {
		return false, err
	}

	if storedOTP != providedOTP {
		return false, nil
	}

	key := fmt.Sprintf("otp:%s", email)
	err = store.Client.Del(store.Ctx, key).Err()
	if err != nil {
		fmt.Printf("Warning: failed to delete OTP: %v\n", err)
	}

	return true, nil
}

// DeleteOTP manually deletes an OTP
func (store *ValkeyChatStore) DeleteOTP(email string) error {
	key := fmt.Sprintf("otp:%s", email)
	return store.Client.Del(store.Ctx, key).Err()
}

// CheckOTPExists checks if an OTP exists for the given email
func (store *ValkeyChatStore) CheckOTPExists(email string) (bool, error) {
	key := fmt.Sprintf("otp:%s", email)
	exists, err := store.Client.Exists(store.Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// GetOTPTTL returns the remaining time-to-live for an OTP
func (store *ValkeyChatStore) GetOTPTTL(email string) (time.Duration, error) {
	key := fmt.Sprintf("otp:%s", email)
	ttl, err := store.Client.TTL(store.Ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return ttl, nil
}
