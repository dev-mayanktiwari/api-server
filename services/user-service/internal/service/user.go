package service

import (
	"errors"
	"fmt"

	"user-service/internal/model"
	"user-service/internal/repository"
	"user-service/pkg/logger"
)

type UserService struct {
	userRepo   *repository.UserRepository
	authClient *AuthClient
	logger     *logger.Logger
}

func NewUserService(userRepo *repository.UserRepository, authClient *AuthClient, logger *logger.Logger) *UserService {
	return &UserService{
		userRepo:   userRepo,
		authClient: authClient,
		logger:     logger,
	}
}

func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.SafeUser, error) {
	if s.userRepo.EmailExists(req.Email, "") {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := &model.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		IsActive:  true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

func (s *UserService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid email or password")
	}

	tokens, err := s.authClient.GenerateTokens(user.ID, user.Email, user.Role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate tokens")
		return nil, fmt.Errorf("failed to generate authentication tokens: %w", err)
	}

	safeUser := user.ToSafeUser()
	return &model.LoginResponse{
		User:         safeUser,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt,
	}, nil
}

func (s *UserService) GetUserByID(userID string) (*model.SafeUser, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

func (s *UserService) UpdateUser(userID string, req *model.UpdateUserRequest) (*model.SafeUser, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Email != "" && req.Email != user.Email {
		if s.userRepo.EmailExists(req.Email, userID) {
			return nil, errors.New("email is already taken")
		}
		user.Email = req.Email
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}

	if req.LastName != "" {
		user.LastName = req.LastName
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

func (s *UserService) ChangePassword(userID string, req *model.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if !user.CheckPassword(req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *UserService) DeleteUser(userID string) error {
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	return s.userRepo.Delete(userID)
}

func (s *UserService) ListUsers(page, limit int) ([]*model.SafeUser, int64, error) {
	offset := (page - 1) * limit
	users, total, err := s.userRepo.List(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var safeUsers []*model.SafeUser
	for _, user := range users {
		safeUser := user.ToSafeUser()
		safeUsers = append(safeUsers, &safeUser)
	}

	return safeUsers, total, nil
}