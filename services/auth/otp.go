package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Request structure for Lambda function
type OTPRequest struct {
	Email string `json:"email"`
}

// Response structure from Lambda function
type OTPResponse struct {
	Success bool   `json:"success"`
	OTP     string `json:"otp,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CallLambdaSendOTP sends email to Lambda function and returns success status and OTP
func CallLambdaSendOTP(email string) (bool, string, error) {
	// Lambda function URL
	url := "https://ge4liwjinjmggwbmwya63lxwt40clppt.lambda-url.ap-south-1.on.aws/"

	// Create request payload
	payload := OTPRequest{
		Email: email,
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var otpResp OTPResponse
	err = json.Unmarshal(body, &otpResp)
	if err != nil {
		return false, "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("lambda returned error: %s", otpResp.Error)
	}

	return otpResp.Success, otpResp.OTP, nil
}

