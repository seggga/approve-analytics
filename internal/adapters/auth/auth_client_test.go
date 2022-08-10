package auth

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/seggga/approve-analytics/internal/domain/models"
)

const (
	AccessToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3QxMjMiLCJpc3MiOiJ0ZWFtOSIsInN1YiI6ImF1dGgtc2VydmljZSIsImV4cCI6MTY1NjgyNTg3N30.HkZnEvcuRsR3l7alzW8THsXre5bKr6Ljlmd94uakovU"
	RefreshToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3QxMjMiLCJpc3MiOiJ0ZWFtOSIsInN1YiI6ImF1dGgtc2VydmljZSIsImV4cCI6MTY1NjgyOTQxN30.7awpddRRrKMwNa6xtl1rQoFaQcXA65ozn9TzzpefnCc"
)

func TestAuthenticate(t *testing.T) {
	var c Client

	t.Run(fmt.Sprintf("Authenticate:%d", http.StatusOK), func(t *testing.T) {
		ctx := context.TODO()

		tokenPair, _ := c.Authenticate(ctx, &models.TokenPair{
			Access:  AccessToken,
			Refresh: RefreshToken,
		})

		if tokenPair.Login != "user123" {
			t.Fatalf("Expected %s, but was %s", "user123", tokenPair.Login)
		}
	})
}
