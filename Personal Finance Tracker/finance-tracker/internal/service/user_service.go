package service

import (
	"errors"
	"personal-finance-tracker/internal/models"
	"personal-finance-tracker/internal/repository"
	"personal-finance-tracker/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(user *models.User) error
	Login(loginReq *models.LoginRequest) (*models.User, string, error)
	GetUserByID(id int) (*models.User, error)
	ValidateToken(token string) (*models.User, error)
	Logout(token string) error
	SetDefaultCurrency(userID, currencyID int) error
	GetUserWithAccounts(userID int) (*models.User, []models.Account, error)
}

type userService struct {
	repo           repository.Repository
	accountService AccountService
}

func NewUserService(repo repository.Repository) UserService {
	accountService := NewAccountService(repo)
	return &userService{
		repo:           repo,
		accountService: accountService,
	}
}

func (s *userService) Register(user *models.User) error {
	// Проверяем, существует ли пользователь с таким email
	existingUser, err := s.repo.GetUserByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("user with this email already exists")
	}

	// Создаем пользователя
	err = s.repo.CreateUser(user)
	if err != nil {
		return err
	}

	// Больше не устанавливаем валюту/счет по умолчанию автоматически.
	// Клиент после первого входа выполнит явный выбор валюты.
	return nil
}

func (s *userService) Login(loginReq *models.LoginRequest) (*models.User, string, error) {
	// Находим пользователя по email
	user, err := s.repo.GetUserByEmail(loginReq.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid email or password")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	// Создаем сессию в БД
	session := &models.Session{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7), // 7 дней
	}

	err = s.repo.CreateSession(session)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *userService) GetUserByID(id int) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *userService) ValidateToken(token string) (*models.User, error) {
	// Проверяем сессию в БД
	session, err := s.repo.GetSessionByToken(token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("invalid or expired token")
	}

	// Получаем пользователя
	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *userService) Logout(token string) error {
	return s.repo.DeleteSession(token)
}

func (s *userService) SetDefaultCurrency(userID, currencyID int) error {
	// Проверяем существование валюты
	currency, err := s.repo.GetCurrencyByID(currencyID)
	if err != nil {
		return err
	}
	if currency == nil {
		return errors.New("currency not found")
	}

	// Устанавливаем валюту по умолчанию пользователю
	if err := s.repo.SetUserDefaultCurrency(userID, currencyID); err != nil {
		return err
	}

	// Если у пользователя нет счетов, создаем дефолтный счет в выбранной валюте
	accounts, err := s.accountService.GetUserAccounts(userID)
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		defaultAccount := &models.Account{
			UserID:     userID,
			CurrencyID: currencyID,
			Balance:    0,
			IsDefault:  true,
		}
		if err := s.accountService.CreateAccount(defaultAccount); err != nil {
			return err
		}
	}

	return nil
}

func (s *userService) GetUserWithAccounts(userID int) (*models.User, []models.Account, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("user not found")
	}

	accounts, err := s.accountService.GetUserAccounts(userID)
	if err != nil {
		return user, nil, err
	}

	return user, accounts, nil
}
