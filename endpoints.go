package wordfeud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CreateAccount creates a new Wordfeud account.
func (c *Client) CreateAccount(username, email, password string) (*Login, SessionID, error) {
	body, err := json.Marshal(struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Username: username,
		Email:    email,
		Password: hashPassword(password),
	})
	if err != nil {
		return nil, "", fmt.Errorf("marshalling request body: %v", err)
	}
	res, err := c.request(http.MethodPost, "/user/create", "", body)
	if err != nil {
		return nil, "", err
	}

	var login Login
	err = json.Unmarshal(res.Content, &login)
	if err != nil {
		return nil, "", fmt.Errorf("unmarshalling response: %v", err)
	}

	sessionID, err := extractSessionID(res)
	if err != nil {
		return nil, "", err
	}

	return &login, sessionID, nil
}

// LoginWithEmail authenticates a user with email and password.
func (c *Client) LoginWithEmail(email, password string) (SessionID, error) {
	body, err := json.Marshal(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    email,
		Password: hashPassword(password),
	})
	if err != nil {
		return "", fmt.Errorf("marshalling request body: %v", err)
	}

	res, err := c.request(http.MethodPost, "/user/login/email", "", body)
	if err != nil {
		return "", err
	}

	return extractSessionID(res)
}

// LoginWithID authenticates a user with id and password.
func (c *Client) LoginWithID(id UserID, password string) (SessionID, error) {
	body, err := json.Marshal(struct {
		ID       UserID `json:"id"`
		Password string `json:"password"`
	}{
		ID:       id,
		Password: hashPassword(password),
	})
	if err != nil {
		return "", fmt.Errorf("marshalling request body: %v", err)
	}

	res, err := c.request(http.MethodPost, "/user/login/id", "", body)
	if err != nil {
		return "", err
	}

	return extractSessionID(res)
}

// ChangePassword changes the password of the user authenticated by session.
func (c *Client) ChangePassword(session SessionID, newPassword string) error {
	body, err := json.Marshal(struct {
		Password string `json:"password"`
	}{hashPassword(newPassword)})
	if err != nil {
		return fmt.Errorf("marshalling request body: %v", err)
	}
	_, err = c.request(http.MethodPost, "/user/password/set", session, body)
	return err
}

// UpdateAvatar updates the avatar of the user authenticated by session and returns the time it was
// updated, as reported by the server.
func (c *Client) UpdateAvatar(session SessionID, image io.Reader) (Timestamp, error) {
	var encoded bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
	_, err := io.Copy(encoder, image)
	if err != nil {
		return Timestamp{}, fmt.Errorf("reading image: %v", err)
	}
	err = encoder.Close()
	if err != nil {
		return Timestamp{}, fmt.Errorf("encoding image to base64: %v", err)
	}

	res, err := roundtrip[struct {
		AvatarUpdated Timestamp `json:"avatar_updated"`
	}](c, http.MethodPost, "/user/avatar/upload", session, struct {
		ImageData string `json:"image_data"`
	}{encoded.String()})
	if err != nil {
		return Timestamp{}, err
	}
	return res.AvatarUpdated, nil
}

// Relationships returns all the friends of the user authenticated by session.
func (c *Client) Relationships(session SessionID) ([]Relationship, error) {
	res, err := roundtrip[struct {
		Relationships []Relationship `json:"relationships"`
	}](c, http.MethodGet, "/user/relationships", session, nil)
	if err != nil {
		return nil, err
	}
	return res.Relationships, nil
}

// CreateRelationship adds a user to the friends list.
func (c *Client) CreateRelationship(session SessionID, user UserID) (*Relationship, error) {
	return roundtrip[Relationship](c, http.MethodPost, "/relationship/create", session, struct {
		ID   UserID `json:"id"`
		Type int    `json:"type"`
	}{ID: user, Type: 0})
}

// DeleteRelationship removes a user from the friends list.
func (c *Client) DeleteRelationship(session SessionID, user UserID) error {
	_, err := c.request(http.MethodPost, fmt.Sprintf("/relationship/%d/delete", user), session, nil)
	return err
}

// Games returns all ongoing games the user authenticated by session is participating in, as well as recently
// finished ones.
func (c *Client) Games(session SessionID) ([]Game, error) {
	res, err := roundtrip[struct {
		Games []Game `json:"games"`
	}](c, http.MethodGet, "/user/games", session, nil)
	if err != nil {
		return nil, err
	}
	return res.Games, nil
}

// Game returns a single game.
func (c *Client) Game(session SessionID, game GameID) (*Game, error) {
	res, err := roundtrip[struct {
		Game Game `json:"game"`
	}](c, http.MethodGet, fmt.Sprintf("/game/%d", game), session, nil)
	if err != nil {
		return nil, err
	}
	return &res.Game, nil
}

// Invite invites a player to a new game by username.
func (c *Client) Invite(session SessionID, username string, ruleset RulesetID, board BoardID) (*Invitation, error) {
	res, err := roundtrip[struct {
		Invitation Invitation `json:"invitation"`
	}](c, http.MethodPost, "/invite/new", session, struct {
		Invitee   string    `json:"invitee"`
		Ruleset   RulesetID `json:"ruleset"`
		BoardType string    `json:"board_type"`
	}{
		Invitee:   username,
		Ruleset:   ruleset,
		BoardType: board.String(),
	})
	if err != nil {
		return nil, err
	}
	return &res.Invitation, nil
}

// InviteRandomOpponent invites a random opponent to a new game.
func (c *Client) InviteRandomOpponent(session SessionID, ruleset RulesetID, board BoardID) (*Invitation, error) {
	res, err := roundtrip[struct {
		Invitation Invitation `json:"invitation"`
	}](c, http.MethodPost, "random_request/create", session, struct {
		Ruleset   RulesetID `json:"ruleset"`
		BoardType string    `json:"board_type"`
	}{
		Ruleset:   ruleset,
		BoardType: board.String(),
	})
	if err != nil {
		return nil, err
	}
	return &res.Invitation, nil
}

// AcceptInvitation accepts a game invitation and returns the id of the resulting game.
func (c *Client) AcceptInvitation(session SessionID, invitation InvitationID) (GameID, error) {
	res, err := roundtrip[struct {
		ID GameID `json:"id"`
	}](c, http.MethodPost, fmt.Sprintf("/invite/%d/accept", invitation), session, nil)
	if err != nil {
		return 0, err
	}
	return res.ID, nil
}

// RejectInvitation rejects a game invitation.
func (c *Client) RejectInvitation(session SessionID, invitation InvitationID) error {
	_, err := c.request(http.MethodPost, fmt.Sprintf("/invite/%d/reject", invitation), session, nil)
	return err
}

// Move performs a move.
func (c *Client) Move(session SessionID, game GameID, move []Placement) (*MoveResult, error) {
	// The API crashes without any actionable error information when attempting to place multiple tiles
	// on the same square. Since this error is pretty hard to debug, we do a check for it here.
	if !uniqueSquares(move) {
		return nil, ErrIllegalMove
	}

	var placements [][4]any
	for _, p := range move {
		placements = append(placements, p.Array())
	}

	return roundtrip[MoveResult](c, http.MethodPost, fmt.Sprintf("/game/%d/move", game), session, struct {
		Move [][4]any `json:"move"`
	}{placements})
}

// Pass passes the turn to the opponent.
func (c *Client) Pass(session SessionID, game GameID) (*MoveResult, error) {
	return roundtrip[MoveResult](c, http.MethodPost, fmt.Sprintf("/game/%d/pass", game), session, nil)
}

// Resign resigns from a game.
func (c *Client) Resign(session SessionID, game GameID) (*MoveResult, error) {
	return roundtrip[MoveResult](c, http.MethodPost, fmt.Sprintf("/game/%d/resign", game), session, nil)
}

// ChatMessages returns all the chat messages sent in a game.
func (c *Client) ChatMessages(session SessionID, game GameID) ([]Message, error) {
	res, err := roundtrip[struct {
		Messages []Message `json:"messages"`
	}](c, http.MethodGet, fmt.Sprintf("/user/%d/chat", game), session, nil)
	if err != nil {
		return nil, err
	}
	return res.Messages, nil
}

// SendChatMessage sends a chat message and returns the time it was sent, as reported by the server.
func (c *Client) SendChatMessage(session SessionID, game GameID, message string) (Timestamp, error) {
	res, err := roundtrip[struct {
		Sent Timestamp `json:"sent"`
	}](c, http.MethodPost, fmt.Sprintf("/game/%d/chat/send", game), session, struct {
		Message string `json:"message"`
	}{message})
	if err != nil {
		return Timestamp{}, err
	}
	return res.Sent, nil
}

// Board returns the layout of a board.
func (c *Client) Board(board BoardID) (*Grid, error) {
	res, err := roundtrip[struct {
		Board Grid `json:"board"`
	}](c, http.MethodGet, fmt.Sprintf("/board/%d", board), "", nil)
	if err != nil {
		return nil, err
	}
	return &res.Board, nil
}
