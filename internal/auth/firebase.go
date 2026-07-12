package auth

import (
    "context"
    "fmt"
    "os"

    firebase "firebase.google.com/go/v4"
    "firebase.google.com/go/v4/auth"
    "google.golang.org/api/option"
)

// Client wraps the Firebase auth client.
// We initialize it once on startup and reuse it — same principle as the DB pool.
type Client struct {
    firebaseAuth *auth.Client
}

func NewClient(ctx context.Context) (*Client, error) {
    credPath := os.Getenv("FIREBASE_CREDENTIALS")
    if credPath == "" {
        return nil, fmt.Errorf("FIREBASE_CREDENTIALS environment variable not set")
    }

    app, err := firebase.NewApp(ctx, nil, option.WithCredentialsFile(credPath))
    if err != nil {
        return nil, fmt.Errorf("error initializing firebase app: %w", err)
    }

    authClient, err := app.Auth(ctx)
    if err != nil {
        return nil, fmt.Errorf("error getting firebase auth client: %w", err)
    }

    return &Client{firebaseAuth: authClient}, nil
}

// VerifyToken validates a Firebase JWT and returns the user's Firebase UID.
// This is called on every authenticated request.
func (c *Client) VerifyToken(ctx context.Context, idToken string) (string, error) {
    token, err := c.firebaseAuth.VerifyIDToken(ctx, idToken)
    if err != nil {
        return "", fmt.Errorf("invalid token: %w", err)
    }
    return token.UID, nil
}