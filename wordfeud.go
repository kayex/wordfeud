package wordfeud

import (
	"encoding/json"
	"math"
	"time"
)

type GameID int64

type Game struct {
	ID            GameID         `json:"id"`
	Board         BoardID        `json:"board"`
	Ruleset       RulesetID      `json:"ruleset"`
	MoveCount     int            `json:"move_count"`
	IsRunning     bool           `json:"is_running"`
	EndGame       int            `json:"end_game"`
	Created       Timestamp      `json:"created"`
	Updated       Timestamp      `json:"updated"`
	CurrentPlayer PlayerPosition `json:"current_player"`
	LastMove      *Move          `json:"last_move"`
	Players       []Player       `json:"players"`
	ChatCount     int            `json:"chat_count"`
	SeenFinished  bool           `json:"seen_finished"`
	ReadChatCount int            `json:"read_chat_count"`
	Hidden        bool           `json:"hidden"`
	Tiles         []Placement    `json:"tiles"`
	BagCount      int            `json:"bag_count"`
	PassCount     int            `json:"pass_count"`
	IsTutorial    bool           `json:"is_tutorial"`
	Rating        *int           `json:"rating"`
	RatingDelta   *int           `json:"rating_delta"`
}

type EndGameStatus int

const (
	EndGameStatusNotFinished EndGameStatus = 0
	EndGameStatusLoss        EndGameStatus = 1
	EndGameStatusWin         EndGameStatus = 2
)

type MoveType string

const (
	MoveTypeMove   MoveType = "move"
	MoveTypePass   MoveType = "pass"
	MoveTypeResign MoveType = "resign"
)

type Move struct {
	MoveType MoveType    `json:"move_type"`
	UserID   UserID      `json:"user_id"`
	Move     []Placement `json:"move"`
	MainWord *string     `json:"main_word"`
	Points   *int        `json:"points"`
}

type Placement struct {
	Column int
	Row    int
	Letter string
	Blank  bool
}

func Place(column int, row int, letter string, blank bool) Placement {
	return Placement{column, row, letter, blank}
}

func (p *Placement) Array() [4]any {
	return [4]any{p.Column, p.Row, p.Letter, p.Blank}
}

func (p *Placement) UnmarshalJSON(b []byte) error {
	var a [4]any
	err := json.Unmarshal(b, &a)
	if err != nil {
		return err
	}

	*p = Placement{
		int(a[0].(float64)),
		int(a[1].(float64)),
		a[2].(string),
		a[3].(bool),
	}
	return nil
}

func uniqueSquares(placements []Placement) bool {
	type position struct {
		column int
		row    int
	}
	positions := make([]position, len(placements))

	for _, pl := range placements {
		for _, pos := range positions {
			if pl.Column == pos.column && pl.Row == pos.row {
				return false
			}
		}
		positions = append(positions, position{column: pl.Column, row: pl.Row})
	}
	return true
}

type Square int

const (
	SquareNormal Square = 0
	SquareDL     Square = 1
	SquareTL     Square = 2
	SquareDW     Square = 3
	SquareTW     Square = 4
)

type Grid [15][15]Square

type BoardID int

const (
	BoardNormal BoardID = 0
	BoardRandom BoardID = 1
)

func (i BoardID) String() string {
	switch i {
	case BoardNormal:
		return "normal"
	case BoardRandom:
		return "random"
	default:
		return ""
	}
}

type Board struct {
	BoardID BoardID `json:"board_id"`
	Board   Grid    `json:"board"`
}

type RulesetID int

const (
	RuleSetAmerican  RulesetID = 0
	RuleSetNorwegian RulesetID = 1
	RuleSetDutch     RulesetID = 2
	RuleSetDanish    RulesetID = 3
	RuleSetSwedish   RulesetID = 4
	RuleSetEnglish   RulesetID = 5
	RuleSetSpanish   RulesetID = 6
	RuleSetFrench    RulesetID = 7
)

type Ruleset struct {
	Ruleset      RulesetID      `json:"ruleset"`
	LanguageCode string         `json:"language_code"`
	TilePoints   map[string]int `json:"tile_points"`
	TileCounts   map[string]int `json:"tile_counts"`
}

type UserID int64

type PlayerPosition int

type Player struct {
	ID            UserID         `json:"id"`
	Username      string         `json:"username"`
	Score         int            `json:"score"`
	Position      PlayerPosition `json:"position"`
	AvatarUpdated Timestamp      `json:"avatar_updated"`
	IsLocal       bool           `json:"is_local"`
	Rack          []string       `json:"rack"`
}

type Login struct {
	Username           string    `json:"username"`
	Email              string    `json:"email"`
	ID                 UserID    `json:"id"`
	Cookies            bool      `json:"cookies"`
	TournamentsEnabled bool      `json:"tournaments_enabled"`
	Created            Timestamp `json:"created"`
	AvatarRoot         string    `json:"avatar_root"`
	AvatarUpdated      Timestamp `json:"avatar_updated"`
	IsGuest            bool      `json:"is_guest"`
}

type Relationship struct {
	UserID        UserID    `json:"user_id"`
	Username      string    `json:"username"`
	AvatarUpdated Timestamp `json:"avatar_updated"`
	Type          int       `json:"type"`
	GamesWon      int       `json:"games_won"`
	GamesLost     int       `json:"games_lost"`
	GamesTied     int       `json:"games_tied"`
}

type Message struct {
	Sent    time.Time `json:"sent"`
	Sender  UserID    `json:"sender"`
	Message string    `json:"message"`
}

type InvitationID int64

type Invitation struct {
	ID        InvitationID `json:"id"`
	Inviter   string       `json:"inviter"`
	InviterID UserID       `json:"inviter_id"`
	Invitee   string       `json:"invitee"`
	InviteeID UserID       `json:"invitee_id"`
	BoardType BoardID      `json:"board_type"`
	Ruleset   RulesetID    `json:"ruleset"`
	Sent      Timestamp    `json:"sent"`
}

type Status struct {
	Games           []GameStatus `json:"games"`
	InvitesSent     []Invitation `json:"invites_sent"`
	InvitesReceived []Invitation `json:"invites_received"`
	RandomRequests  []Invitation `json:"random_requests"`
}

type GameStatus struct {
	ID            GameID    `json:"id"`
	ChatCount     int       `json:"chat_count"`
	Updated       Timestamp `json:"updated"`
	ReadChatCount int       `json:"read_chat_count"`
}

type MoveResult struct {
	Points    *int      `json:"points"`
	MainWord  *string   `json:"main_word"`
	NewTiles  []string  `json:"new_tiles"`
	IsRunning bool      `json:"is_running"`
	Updated   Timestamp `json:"updated"`
	Game      Game      `json:"game"`
}

type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var f float64
	err := json.Unmarshal(b, &f)
	if err != nil {
		return err
	}
	sec, dec := math.Modf(f)
	*t = Timestamp{time.Unix(int64(sec), int64(dec*(1e9)))}
	return nil
}
