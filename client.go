package wordfeud

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Client is a Wordfeud API client. It is safe for concurrent use by multiple goroutines.
type Client struct {
	cl      *http.Client
	baseURL string
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		cl:      http.DefaultClient,
		baseURL: "https://game.wordfeud.com/wf",
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithHTTPClient sets which http.Client is used to perform requests against the API.
func WithHTTPClient(cl *http.Client) ClientOption {
	return func(c *Client) {
		c.cl = cl
	}
}

// WithBaseURL sets the base URL for the Wordfeud API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(baseURL, "/")
	}
}

// SessionID is a Wordfeud authentication session identifier.
type SessionID string

func (s SessionID) cookie() string {
	return fmt.Sprintf("sessionid=%s", s)
}

// PasswordSalt is the salt that is added to the end of all user passwords before being hashed
// and sent to the API.
var PasswordSalt = "JarJarBinks9"

func hashPassword(password string) string {
	h := sha1.New()
	h.Write([]byte(password))
	h.Write([]byte(PasswordSalt))
	return hex.EncodeToString(h.Sum(nil))
}

var ErrAccessDenied = errors.New("access_denied")
var ErrAlreadyExists = errors.New("already_exists")
var ErrDuplicateInvite = errors.New("duplicate_invite")
var ErrGameOver = errors.New("game_over")
var ErrIllegalMove = errors.New("illegal_move")
var ErrIllegalTiles = errors.New("illegal_tiles")
var ErrIllegalUserSelf = errors.New("illegal_user_self")
var ErrIllegalWord = errors.New("illegal_word")
var ErrInvalidBoardType = errors.New("invalid_board_type")
var ErrInvalidID = errors.New("invalid_id")
var ErrInvalidRuleset = errors.New("invalid_ruleset")
var ErrLoginRequired = errors.New("login_required")
var ErrNotFound = errors.New("not_found")
var ErrNotYourTurn = errors.New("not_your_turn")
var ErrUnknownEmail = errors.New("unknown_email")
var ErrUserNotFound = errors.New("user_not_found")
var ErrWrongPassword = errors.New("wrong_password")

func convertToSentinel(e *apiError) error {
	m := map[string]error{
		"access_denied":      ErrAccessDenied,
		"already_exists":     ErrAlreadyExists,
		"duplicate_invite":   ErrDuplicateInvite,
		"game_over":          ErrGameOver,
		"illegal_move":       ErrIllegalMove,
		"illegal_tiles":      ErrIllegalTiles,
		"illegal_user_self":  ErrIllegalUserSelf,
		"illegal_word":       ErrIllegalWord,
		"invalid_board_type": ErrInvalidBoardType,
		"invalid_id":         ErrInvalidID,
		"invalid_ruleset":    ErrInvalidRuleset,
		"login_required":     ErrLoginRequired,
		"not_found":          ErrNotFound,
		"not_your_turn":      ErrNotYourTurn,
		"unknown_email":      ErrUnknownEmail,
		"user_not_found":     ErrUserNotFound,
		"wrong_password":     ErrWrongPassword,
	}

	if s, ok := m[e.Type]; ok {
		return s
	}
	return e
}
