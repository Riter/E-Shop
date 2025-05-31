package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	jwt_tok "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid appID")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login users and returns token
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "Auth.Login"
	log := a.log.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attemping to login user")

	user, err := a.usrProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

// RegisterNewUser registers new user in system and returns user ID
func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	//panic("not inplemented")

	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))

		return 0, fmt.Errorf("%s : %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

// Checkes if user is admin. Return bool value
func (a *Auth) IsAdmin(
	ctx context.Context, userID int64,
) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *Auth) ValidateToken(ctx context.Context, tokenString string) (int64, error) {
	const op = "auth.ValidateToken"

	// 1. Распарсить токен без валидации, чтобы вытащить app_id
	parser := jwt_tok.NewParser(jwt_tok.WithoutClaimsValidation())
	unverifiedToken, _, err := parser.ParseUnverified(tokenString, jwt_tok.MapClaims{})
	if err != nil {
		a.log.Error("failed to parse token", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	claims, ok := unverifiedToken.Claims.(jwt_tok.MapClaims)
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	appIDFloat, ok := claims["app_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}
	appID := int(appIDFloat)

	// 2. Получить App по app_id
	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		a.log.Error("failed to get app by ID", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
	}

	// 3. Провалидировать токен с подписью
	validatedToken, err := jwt_tok.Parse(tokenString, func(t *jwt_tok.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt_tok.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: unexpected signing method", op)
		}
		return []byte(app.Secret), nil
	})
	if err != nil || !validatedToken.Valid {
		a.log.Warn("invalid token", sl.Err(err))
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	validClaims, ok := validatedToken.Claims.(jwt_tok.MapClaims)
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	// 4. Проверка срока действия
	if expRaw, ok := validClaims["exp"].(float64); ok {
		if int64(expRaw) < time.Now().Unix() {
			return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
		}
	}

	// 5. Получение uid
	uidFloat, ok := validClaims["uid"].(float64)
	if !ok {
		return 0, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	return int64(uidFloat), nil
}
