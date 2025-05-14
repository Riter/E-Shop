package auth

import (
	"context"
	"fmt"
	"log/slog"

	ssov1 "github.com/GGiovanni9152/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client ssov1.AuthClient
	log    *slog.Logger
}

func New(addr string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &Client{
		client: ssov1.NewAuthClient(conn),
		log:    log,
	}, nil
}

// IsAdmin проверяет является ли пользователь администратором
func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	resp, err := c.client.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check admin status: %w", err)
	}

	return resp.GetIsAdmin(), nil
}
