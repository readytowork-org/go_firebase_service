package go_firebase_service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go"

	"firebase.google.com/go/auth"
)

type AuthUser struct {
	Password    string
	Role        string
	DisplayName *string
	Email       string
	AdminID     int64
	UserID      int64
}

type authConfigLogger interface {
	Fatalf(template string, args ...interface{})
}

type AuthConfig struct {
	Logger authConfigLogger
	App    *firebase.App
}

// AuthService structure
type AuthService struct {
	*auth.Client
}

// AuthErrorResponse structure
type AuthErrorResponse struct {
	Message   string `json:"message"`
	ErrorType int    `json:"error_type"`
}

type claimsMap map[string]interface{}

// NewFirebaseAuthService creates new firebase service
func NewFirebaseAuthService(config AuthConfig) AuthService {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	firebaseAuth, err := config.App.Auth(ctx)
	if err != nil {
		config.Logger.Fatalf("Firebase Authentication: %v", err)
	}

	return AuthService{
		Client: firebaseAuth,
	}
}

// Create creates a new user with email and password
func (fb *AuthService) Create(
	userRequest AuthUser, setClaims ...func(claims claimsMap) claimsMap,
) (
	string, *AuthErrorResponse,
) {

	params := (&auth.UserToCreate{}).
		Email(userRequest.Email).
		Password(userRequest.Password)

	if userRequest.DisplayName != nil && *userRequest.DisplayName != "" {
		params = params.DisplayName(*userRequest.DisplayName)
	}

	u, err := fb.Client.CreateUser(context.Background(), params)
	if err != nil {
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			errMessage := fmt.Sprintf("%s with this email already exits in Firebase", userRequest.Role)
			return "", &AuthErrorResponse{
				ErrorType: http.StatusBadRequest,
				Message:   errMessage,
			}
		}
		errMsg := fmt.Sprintf("Failed to create %v\n Error :: %v", userRequest, err.Error())
		return "", &AuthErrorResponse{
			ErrorType: http.StatusInternalServerError,
			Message:   errMsg,
		}
	}

	claims := claimsMap{"role": userRequest.Role}

	for _, setClaim := range setClaims {
		claims = setClaim(claims)
	}

	errResponse := fb.SetClaim(u.UID, claims)
	if errResponse != nil {
		return "", &AuthErrorResponse{
			Message:   errResponse.Message,
			ErrorType: http.StatusInternalServerError,
		}
	}
	return u.UID, nil
}

// CreateUser creates a new user with email and password
func (fb *AuthService) CreateUser(userRequest AuthUser) (string, *AuthErrorResponse) {
	return fb.Create(
		userRequest, func(claims claimsMap) claimsMap {
			claims[Claims.UserId.Name()] = userRequest.UserID
			return claims
		},
	)
}

// CreateAdmin creates a new admin with email and password
func (fb *AuthService) CreateAdmin(userRequest AuthUser) (string, *AuthErrorResponse) {
	return fb.Create(
		userRequest, func(claims claimsMap) claimsMap {
			if userRequest.Role != "admin" {
				claims[Claims.AdminId.ToString()] = userRequest.AdminID
			}
			return claims
		},
	)
}

// VerifyToken verify passed firebase id token
func (fb *AuthService) VerifyToken(idToken string) (*auth.Token, *AuthErrorResponse) {
	token, err := fb.VerifyIDToken(context.Background(), idToken)
	if err != nil {
		return nil, &AuthErrorResponse{
			Message: err.Error(),
		}
	}
	return token, nil
}

// SetClaim set's claim to firebase user
func (fb *AuthService) SetClaim(uid string, claims map[string]interface{}) *AuthErrorResponse {
	err := fb.SetCustomUserClaims(context.Background(), uid, claims)
	if err != nil {
		return &AuthErrorResponse{
			Message: err.Error(),
		}
	}
	return nil
}

// UpdateEmailVerification update firebase user email verify
func (fb *AuthService) UpdateEmailVerification(uid string) *AuthErrorResponse {
	params := (&auth.UserToUpdate{}).
		EmailVerified(true)
	_, err := fb.UpdateUser(context.Background(), uid, params)
	if err != nil {
		return &AuthErrorResponse{
			Message: err.Error(),
		}
	}
	return nil
}

// DisableUser true for disabled; false for enabled.
func (fb *AuthService) DisableUser(uid string, disable bool) *AuthErrorResponse {
	params := (&auth.UserToUpdate{}).
		Disabled(disable)
	_, err := fb.UpdateUser(context.Background(), uid, params)
	if err != nil {
		return &AuthErrorResponse{
			Message: err.Error(),
		}
	}
	return nil
}

// UpdateFirebaseAdmin handles the common operation to update admin in Firebase for OneStore Admin and Admin
func (fb *AuthService) UpdateFirebaseAdmin(UID string, newUserData, oldUserData AuthUser) *AuthErrorResponse {
	fbAdmin := &auth.UserToUpdate{}

	if newUserData.Email != "" && newUserData.Email != oldUserData.Email {
		fbAdmin = fbAdmin.Email(newUserData.Email)
	}

	if newUserData.Password != "" {
		fbAdmin = fbAdmin.Password(newUserData.Password)
	}

	if newUserData.DisplayName != nil && newUserData.DisplayName != oldUserData.DisplayName {
		fbAdmin = fbAdmin.DisplayName(*newUserData.DisplayName)
	}

	if fbAdmin != nil {
		if _, err := fb.UpdateUser(context.Background(), UID, fbAdmin); err != nil {
			return &AuthErrorResponse{
				Message: err.Error(),
			}
		}
	}
	return nil
}

func (fb *AuthService) CreateCustomToken(ctx context.Context, uid string) (string, *AuthErrorResponse) {
	token, err := fb.CustomToken(ctx, uid)
	if err != nil {
		return "", &AuthErrorResponse{
			Message: err.Error(),
		}
	}
	return token, nil
}
