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

	"github.com/prometheus/client_golang/prometheus"
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
	ErrUserNotFound       = errors.New("user not found")
)

var (
	loginAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_attempts_total",
		Help: "Total login attempts",
	})

	loginSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_success_total",
		Help: "Total successful logins",
	})

	loginFailures = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_failure_total",
		Help: "Total failed logins",
	})

	loginBadToken = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_bad_token",
		Help: "Total failed token creations",
	})

	loginDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "auth_login_duration_seconds",
		Help:    "Duration of login attempts",
		Buckets: prometheus.DefBuckets,
	})

	registerAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_register_attempts_total",
		Help: "Total register attempts",
	})

	registerFailure = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_register_failure_total",
		Help: "Total failed registerations",
	})

	registerSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_register_success_total",
		Help: "Total successful registrations",
	})

	registerDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "auth_register_duration_seconds",
		Help:    "Durationg of registrations attempts",
		Buckets: prometheus.DefBuckets,
	})

	isadminAttempts = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_isadmin_attempts_total",
		Help: "Total checking if user is admin",
	})

	isadminErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_isadmin_errors_total",
		Help: "Total errors while checking if admin",
	})

	isadminSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_isadmin_success_total",
		Help: "Total successful checks if admin",
	})
)

func init() {
	prometheus.MustRegister(loginAttempts, loginSuccess, loginFailures, loginDuration, loginBadToken,
		registerAttempts, registerDuration, registerFailure, registerSuccess,
		isadminAttempts, isadminErrors, isadminSuccess)
}

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

	timer := prometheus.NewTimer(loginDuration)
	defer timer.ObserveDuration()
	loginAttempts.Inc()

	user, err := a.usrProvider.User(ctx, email)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			loginFailures.Inc()
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))
		loginFailures.Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		loginFailures.Inc()

		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		loginFailures.Inc()
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)

	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		loginBadToken.Inc()

		return "", fmt.Errorf("%s: %w", op, err)
	}

	loginSuccess.Inc()
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

	timer := prometheus.NewTimer(registerDuration)
	defer timer.ObserveDuration()
	registerAttempts.Inc()

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
		registerFailure.Inc()
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))
			registerFailure.Inc()
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))
		registerFailure.Inc()
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	log.Info("user registered")
	registerSuccess.Inc()

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

	isadminAttempts.Inc()

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))
			isadminErrors.Inc()
			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}
		isadminErrors.Inc()
		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))
	isadminSuccess.Inc()
	return isAdmin, nil
}
