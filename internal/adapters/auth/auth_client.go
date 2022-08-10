package auth

import (
	"context"
	"fmt"
	"time"

	pb "github.com/seggga/approve-analytics/internal/adapters/auth/proto"
	"github.com/seggga/approve-analytics/internal/domain/models"
	"google.golang.org/grpc"
)

// Client sends auth requests to AUTH service via gRPC
type Client struct {
	Conn *grpc.ClientConn
	pb.AuthAPIClient
}

// NewClient creates auth Client
func NewClient(addr string) (*Client, error) {
	// create connection
	cwt, _ := context.WithTimeout(context.Background(), time.Second*5)
	conn, err := grpc.DialContext(cwt, addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("error creating connection %s: %v", addr, err)
	}

	// create client
	client := pb.NewAuthAPIClient(conn)

	return &Client{
		Conn:          conn,
		AuthAPIClient: client,
	}, nil
}

// Authenticate ...
func (c *Client) Authenticate(ctx context.Context, tokenPair *models.TokenPair) (*models.TokenPair, error) {
	// create request with access token
	tokenReq := &pb.CheckTokenRequest{
		Token: tokenPair.Access,
	}
	// check access token
	tokenResp, err := c.CheckToken(ctx, tokenReq)
	if err != nil {
		return nil, fmt.Errorf("cannot check access token: %v", err)
	}

	// got a bad error while checking access token
	if tokenResp.GetError() != pb.TokenNotValid && tokenResp.Error != pb.NoError {
		return nil, fmt.Errorf("check token returned error: %v", tokenResp.GetError())
	}
	// access token is valid
	if tokenResp.Error == pb.NoError {
		return &models.TokenPair{
			Login:     tokenResp.GetLogin(),
			Refreshed: false,
		}, nil
	}

	// access token is not valid, we should check refresh token
	tokenReq = &pb.CheckTokenRequest{
		Token: tokenPair.Refresh,
	}
	tokenResp, err = c.CheckToken(ctx, tokenReq)
	if err != nil {
		return nil, fmt.Errorf("cannot check refresh token: %v", err)
	}

	// got a bad error while checking refresh token
	if tokenResp.GetError() != pb.NoError {
		return nil, fmt.Errorf("Unauthorized: %v", tokenResp.GetError())
	}

	refreshReq := &pb.RefreshTokensRequest{
		Token: tokenPair.Refresh,
	}
	refreshResp, err := c.RefreshTokens(ctx, refreshReq)
	if err != nil {
		return nil, fmt.Errorf("cannot check refresh token: %v", err)
	}
	// got a bad error while checking refresh token
	if refreshResp.GetError() != pb.NoError {
		return nil, fmt.Errorf("Unauthorized: %v", refreshResp.GetError())
	}

	return &models.TokenPair{
		Access:    refreshResp.GetAccessToken(),
		Refresh:   refreshResp.GetRefreshToken(),
		Login:     tokenResp.GetLogin(),
		Refreshed: true,
	}, nil
}
