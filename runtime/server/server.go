package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/creack/ehttp"
	"github.com/creack/gofour/engine"
	"github.com/creack/gofour/runtime"
	"github.com/creack/httpreq"
	"github.com/creack/uuid"
)

func init() {
	runtime.Runtimes["server"] = &Runtime{}
}

// Runtime is a HTTP server for Connect Four.
type Runtime struct {
	sync.RWMutex
	games map[string]*engine.Four
}

// CreateGameReq is the request to create a new game.
type CreateGameReq struct {
	Cols     int
	Rows     int
	NPlayers int
	NWin     int
}

// CreateGame is the http endpoint handling the game creation.
//
// Method: GET
// Query String:
// - cols:     int, columns count of the grid.
// - rows:     int, rows count of the grid.
// - nplayers: int, number of players allowed in the game.
// - nwin:     int, number of consecutive field to win.
// Response:
// - json formatted UUID of the new game.
func (r *Runtime) CreateGame(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}
	data := &CreateGameReq{
		Cols:     engine.DefaultCols,
		Rows:     engine.DefaultRows,
		NPlayers: engine.DefaultNPlayers,
		NWin:     engine.DefaultNWin,
	}
	if err := (httpreq.ParsingMap{
		{Field: "cols", Fct: httpreq.ToInt, Dest: &data.Cols},
		{Field: "rows", Fct: httpreq.ToInt, Dest: &data.Rows},
		{Field: "nplayers", Fct: httpreq.ToInt, Dest: &data.NPlayers},
		{Field: "nwin", Fct: httpreq.ToInt, Dest: &data.NWin},
	}.Parse(req.Form)); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}

	four, err := engine.NewConnectFour(data.Cols, data.Rows, data.NPlayers, data.NWin)
	if err != nil {
		return ehttp.NewErrorf(http.StatusInternalServerError, "error instantiating new game: %s", err)
	}

	gameID := uuid.New()
	r.Lock()
	r.games[gameID] = four
	r.Unlock()

	return json.NewEncoder(w).Encode(gameID)
}

// ListGameResp is the response
type ListGameResp struct {
	GameID         string                  `json:"game_id"`
	PlayerCount    int                     `json:"player_count"`
	MaxPlayerCount int                     `json:"max_player_count"`
	GameState      string                  `json:"game_state"`
	Players        map[engine.State]string `json:"players"`
}

// ListGames is the http endpoint returning the list of games.
//
// Method: GET
// Response:
// - JSON array of ListGameResp.
func (r *Runtime) ListGames(w http.ResponseWriter, req *http.Request) error {
	r.RLock()
	ret := make([]ListGameResp, 0, len(r.games))
	for gameID, game := range r.games {
		gameState := "pending"
		if game.GridState == engine.Stale {
			gameState = "stale"
		} else if game.GridState != engine.Empty {
			gameState = fmt.Sprintf("won by %s %s", game.Players[game.GridState], game.GridState)
		}
		ret = append(ret, ListGameResp{
			GameID:         gameID,
			PlayerCount:    len(game.Players),
			MaxPlayerCount: game.NPlayers,
			GameState:      gameState,
			Players:        game.Players,
		})
	}
	r.RUnlock()
	return json.NewEncoder(w).Encode(ret)
}

// AttachGame is the http endpoint to attach to a game.
// This endpoint will send one message each time the game changes state until
// the game is finished.
//
// Method: GET
// Query String:
// - game_id: string, game uuid to attach to.
// Response:
// - JSON object of engine.Four. One entry per state change.
func (r *Runtime) AttachGame(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}
	gameID := req.Form.Get("game_id")
	if gameID == "" {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing game_id")
	}
	if uuid.Parse(gameID) == nil {
		return ehttp.NewErrorf(http.StatusBadRequest, "invalid game_id")
	}
	r.RLock()
	game := r.games[gameID]
	r.RUnlock()
	if game == nil {
		return ehttp.NewErrorf(http.StatusNotFound, "game '%s' not found", gameID)
	}

	// Send current state.
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(game); err != nil {
		return ehttp.NewError(http.StatusInternalServerError, err)
	}
	w.(http.Flusher).Flush()

	// For each state change, resend the game.
	if game.ActivityChan == nil {
		return nil
	}
	for range game.ActivityChan {
		if err := encoder.Encode(game); err != nil {
			return ehttp.NewError(http.StatusInternalServerError, err)
		}
		w.(http.Flusher).Flush()
	}
	return nil
}

// JoinGameReq is the request to join a game.
type JoinGameReq struct {
	GameID     string
	PlayerName string
}

// JoinGame is the http endpoint to join a game.
//
// Method: GET
// Query String:
// - game_id:     string, uuid of the game to join.
// - player_name: string, arbitrary player name.
func (r *Runtime) JoinGame(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}
	data := &JoinGameReq{}
	if err := (httpreq.ParsingMap{
		{Field: "game_id", Fct: httpreq.ToString, Dest: &data.GameID},
		{Field: "player_name", Fct: httpreq.ToString, Dest: &data.PlayerName},
	}.Parse(req.Form)); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}
	if data.PlayerName == "" {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing player name")
	}
	if data.GameID == "" {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing game id")
	}
	if uuid.Parse(data.GameID) == nil {
		return ehttp.NewErrorf(http.StatusBadRequest, "invalid game id format")
	}
	r.RLock()
	game := r.games[data.GameID]
	r.RUnlock()
	if game == nil {
		return ehttp.NewErrorf(http.StatusNotFound, "game '%s' not found", data.GameID)
	}
	if len(game.Players) >= game.NPlayers {
		return ehttp.NewErrorf(http.StatusForbidden, "game '%s' is full", data.GameID)
	}
	for _, playerName := range game.Players {
		if playerName == data.PlayerName {
			return ehttp.NewErrorf(http.StatusForbidden, "user '%s' already joined game '%s'", data.PlayerName, data.GameID)
		}
	}
	game.Lock()
	game.Players[engine.State(len(game.Players)+1)] = data.PlayerName
	game.Unlock()

	return nil
}

// PlayMoveReq is the request to play a move in a game.
type PlayMoveReq struct {
	GameID     string
	PlayerName string
	Column     int
}

// PlayMove is the http endpoint to submit a move.
//
// Method: GET
// Query String:
// - game_id:     string, uuid of the target game.
// - player_name: string, name of the player, must have joined the game.
// - col:         int,    0 indexed column number to play.
func (r *Runtime) PlayMove(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}

	data := PlayMoveReq{
		Column: -1,
	}
	if err := (httpreq.ParsingMap{
		{Field: "game_id", Fct: httpreq.ToString, Dest: &data.GameID},
		{Field: "player_name", Fct: httpreq.ToString, Dest: &data.PlayerName},
		{Field: "col", Fct: httpreq.ToInt, Dest: &data.Column},
	}.Parse(req.Form)); err != nil {
		return ehttp.NewError(http.StatusBadRequest, err)
	}
	if data.Column == -1 {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing column to be played")
	}
	if data.PlayerName == "" {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing player name")
	}
	if data.GameID == "" {
		return ehttp.NewErrorf(http.StatusBadRequest, "missing game id")
	}
	if uuid.Parse(data.GameID) == nil {
		return ehttp.NewErrorf(http.StatusBadRequest, "invalid game id format")
	}
	r.RLock()
	game := r.games[data.GameID]
	r.RUnlock()
	if game == nil {
		return ehttp.NewErrorf(http.StatusNotFound, "game '%s' not found", data.GameID)
	}
	if len(game.Players) != game.NPlayers {
		return ehttp.NewErrorf(http.StatusForbidden, "game '%s' is not ready, waiting on players", data.GameID)
	}
	var player engine.State
	for s, playerName := range game.Players {
		if playerName == data.PlayerName {
			player = s
			break
		}
	}
	if player == engine.Empty {
		return ehttp.NewErrorf(http.StatusForbidden, "player not found in game '%s'", data.GameID)
	}
	if err := game.ValidateMove(player, data.Column); err != nil {
		return ehttp.NewErrorf(http.StatusForbidden, "invalid move for player '%s' in game '%s': %s", data.PlayerName, data.GameID, err)
	}
	if _, err := game.PlayerMove(player, data.Column); err != nil {
		return ehttp.NewErrorf(http.StatusInternalServerError, "an error occured while playing a move: %s", err)
	}
	return nil
}

// Init setup the connect four game.
// Note: In server mode, we discard the init's given engine.
func (r *Runtime) Init(four *engine.Four) error {
	r.games = map[string]*engine.Four{}

	ehttp.HandleFunc("/create", ehttp.HandlerFunc(r.CreateGame))
	ehttp.HandleFunc("/join", ehttp.HandlerFunc(r.JoinGame))
	ehttp.HandleFunc("/list", ehttp.HandlerFunc(r.ListGames))
	ehttp.HandleFunc("/attach", ehttp.HandlerFunc(r.AttachGame))
	ehttp.HandleFunc("/play", ehttp.HandlerFunc(r.PlayMove))
	return nil
}

// Run is the main loop.
// TODO: Use flags to config the listen address.
func (r *Runtime) Run() error {
	return http.ListenAndServe("0.0.0.0:8080", nil)
}

// Close .
func (r *Runtime) Close() error {
	return nil
}
