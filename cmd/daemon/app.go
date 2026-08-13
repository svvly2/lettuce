package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/svvly2/lettuce/internal/app/assets"
	appconfig "github.com/svvly2/lettuce/internal/app/config"
	"github.com/svvly2/lettuce/internal/app/request"
	"github.com/svvly2/lettuce/internal/app/response"
	"github.com/svvly2/lettuce/internal/files"
	"github.com/svvly2/lettuce/internal/oauth"
	"github.com/svvly2/lettuce/internal/roblox"
)

// LogEntry is a single line in the activity log.
type LogEntry struct {
	ID      int64  `json:"id"`
	Level   string `json:"level"` // "info" | "success" | "error" | "warn"
	Message string `json:"message"`
	Time    string `json:"time"`
}

type UploadJob struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	SourceAssetID string `json:"sourceAssetId"`
	ResultAssetID string `json:"resultAssetId,omitempty"`
	Progress      int    `json:"progress"`
	Attempt       int    `json:"attempt"`
	Status        string `json:"status"`
	finishedAt    time.Time
}

const terminalJobRetention = 30 * time.Second

// AppState holds all application state shared between screens.
type AppState struct {
	mu sync.Mutex

	// Identity
	IsLoggedIn   bool
	RefreshToken string
	OAuthManager *oauth.Manager

	// Roblox upload client.
	Client *roblox.Client

	// Server
	ServerRunning bool
	ServerPort    string
	serverStop    chan struct{}
	Busy          bool
	finished      bool

	// Responses for JSON export
	resp        *response.Response
	respHistory []response.ResponseItem
	exportJSON  bool
	exportName  string

	// Live log
	Log     []LogEntry
	LogSeq  int64
	Queue   []UploadJob
	OnLog   func(LogEntry) // called on every new entry (triggers UI refresh)
	OnState func()         // called when auth/server state changes

	cookieInputChan chan string
}

func newAppState() *AppState {
	s := &AppState{
		ServerPort:      appconfig.Get("port"),
		finished:        true,
		resp:            response.New(),
		cookieInputChan: make(chan string, 1),
	}
	sessionFile := appconfig.Get("oauth_session_file")
	if sessionFile == "" {
		sessionFile = "oauth-session.json"
	}
	s.OAuthManager = oauth.NewManager(sessionFile)
	s.OAuthManager.SetOnSessionUpdate(s.applyOAuthSession)
	return s
}

// AddLog appends a log entry and calls the OnLog callback on the UI goroutine.
func (s *AppState) AddLog(level, msg string) {
	s.LogSeq++
	entry := LogEntry{
		ID:      s.LogSeq,
		Level:   level,
		Message: msg,
		Time:    time.Now().Format("15:04:05"),
	}
	s.mu.Lock()
	s.Log = append(s.Log, entry)
	cb := s.OnLog
	s.mu.Unlock()
	if cb != nil {
		cb(entry)
	}
}

type StateSnapshot struct {
	IsLoggedIn    bool        `json:"isLoggedIn"`
	UserID        int64       `json:"userId"`
	Username      string      `json:"username"`
	DisplayName   string      `json:"displayName"`
	AvatarURL     string      `json:"avatarUrl"`
	ServerRunning bool        `json:"serverRunning"`
	ServerPort    string      `json:"serverPort"`
	Busy          bool        `json:"busy"`
	Logs          []LogEntry  `json:"logs"`
	Queue         []UploadJob `json:"queue"`
}

func (s *AppState) Snapshot() StateSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	var username, displayName string
	var userID int64
	var avatarURL string
	if s.Client != nil {
		userID = s.Client.UserInfo.ID
		username = s.Client.UserInfo.Username
		displayName = s.Client.UserInfo.DisplayName
		avatarURL = s.Client.UserInfo.Picture
	}
	logs := make([]LogEntry, len(s.Log))
	copy(logs, s.Log)
	queue := make([]UploadJob, len(s.Queue))
	copy(queue, s.Queue)
	return StateSnapshot{
		IsLoggedIn:    s.IsLoggedIn,
		UserID:        userID,
		Username:      username,
		DisplayName:   displayName,
		AvatarURL:     avatarURL,
		ServerRunning: s.ServerRunning,
		ServerPort:    s.ServerPort,
		Busy:          s.Busy,
		Logs:          logs,
		Queue:         queue,
	}
}

func (s *AppState) startJobs(ids []int64) {
	s.mu.Lock()
	s.Queue = make([]UploadJob, len(ids))
	for i, id := range ids {
		value := strconv.FormatInt(id, 10)
		s.Queue[i] = UploadJob{ID: value, Name: "Animation " + value, SourceAssetID: value, Progress: 8, Attempt: 1, Status: "uploading"}
	}
	s.mu.Unlock()
}

func (s *AppState) completeJob(oldID, newID int64) {
	s.mu.Lock()
	source := strconv.FormatInt(oldID, 10)
	var finishedAt time.Time
	for i := range s.Queue {
		if s.Queue[i].SourceAssetID == source {
			s.Queue[i].ResultAssetID = strconv.FormatInt(newID, 10)
			s.Queue[i].Progress = 100
			s.Queue[i].Status = "complete"
			finishedAt = time.Now()
			s.Queue[i].finishedAt = finishedAt
			break
		}
	}
	s.mu.Unlock()
	if !finishedAt.IsZero() {
		go s.removeFinishedJobAfter(source, finishedAt, terminalJobRetention)
	}
}

func (s *AppState) finishJobs() {
	s.mu.Lock()
	type finishedJob struct {
		source     string
		finishedAt time.Time
	}
	finished := make([]finishedJob, 0)
	for i := range s.Queue {
		if s.Queue[i].Status != "complete" && s.Queue[i].Status != "failed" {
			s.Queue[i].Progress = 100
			s.Queue[i].Status = "failed"
			s.Queue[i].finishedAt = time.Now()
			finished = append(finished, finishedJob{s.Queue[i].SourceAssetID, s.Queue[i].finishedAt})
		}
	}
	s.mu.Unlock()
	for _, job := range finished {
		go s.removeFinishedJobAfter(job.source, job.finishedAt, terminalJobRetention)
	}
}

func (s *AppState) removeFinishedJobAfter(source string, finishedAt time.Time, delay time.Duration) {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	<-timer.C

	s.mu.Lock()
	kept := s.Queue[:0]
	removed := false
	for _, job := range s.Queue {
		if job.SourceAssetID == source && job.finishedAt.Equal(finishedAt) && (job.Status == "complete" || job.Status == "failed") {
			removed = true
			continue
		}
		kept = append(kept, job)
	}
	s.Queue = kept
	cb := s.OnState
	s.mu.Unlock()
	if removed && cb != nil {
		cb()
	}
}

func (s *AppState) LogsSince(id int64) []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]LogEntry, 0)
	for _, entry := range s.Log {
		if entry.ID > id {
			out = append(out, entry)
		}
	}
	return out
}

// tryRestoreSession is intentionally a no-op for OAuth login.
func (s *AppState) tryRestoreSession() {
	if s.OAuthManager == nil {
		return
	}
	session := s.OAuthManager.CurrentSession()
	if session == nil {
		return
	}
	if !session.IsValid() {
		var err error
		session, err = s.OAuthManager.RefreshSessionIfNeeded()
		if err != nil {
			_ = s.OAuthManager.ClearSession()
			return
		}
	}
	// Sessions created before profile scope was enabled cannot render identity.
	// Clear once and let the user authorize the corrected scope set.
	if strings.TrimSpace(session.Username) == "" {
		_ = s.OAuthManager.ClearSession()
		return
	}
	s.mu.Lock()
	s.Client = roblox.NewOAuthAuthenticatedClient(session.AccessToken, roblox.UserInfo{ID: session.UserID, Username: session.Username, DisplayName: session.DisplayName, Picture: session.Picture})
	s.RefreshToken = session.RefreshToken
	s.IsLoggedIn = true
	s.mu.Unlock()
}

// Logout clears credentials and resets auth state.
func (s *AppState) Logout() {
	if s.OAuthManager != nil {
		_ = s.OAuthManager.ClearSession()
	}
	s.mu.Lock()
	s.IsLoggedIn = false
	s.Client = nil
	s.RefreshToken = ""
	cb := s.OnState
	s.mu.Unlock()
	if cb != nil {
		cb()
	}
}

// SetAPIKey stores the Open Cloud API key in the keyring and config.
func (s *AppState) SetAPIKey(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	appconfig.Set("api_key", key)
	if err := appconfig.PersistAPIKey(); err != nil {
		return fmt.Errorf("save API key: %w", err)
	}
	s.AddLog("success", "API key saved.")
	return nil
}

// StartServer launches the local HTTP server.
func (s *AppState) StartServer() error {
	s.mu.Lock()
	if s.ServerRunning {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	port := s.ServerPort
	stopCh := make(chan struct{})
	s.serverStop = stopCh
	s.mu.Unlock()

	mux := http.NewServeMux()
	s.setupRoutes(mux)

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		s.AddLog("info", fmt.Sprintf("Server listening on localhost:%s â€” waiting for pluginâ€¦", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.AddLog("error", "Server error: "+err.Error())
		}
	}()

	go func() {
		<-stopCh
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		s.mu.Lock()
		s.ServerRunning = false
		cb := s.OnState
		s.mu.Unlock()
		if cb != nil {
			cb()
		}
		s.AddLog("info", "Server stopped.")
	}()

	s.mu.Lock()
	s.ServerRunning = true
	cb := s.OnState
	s.mu.Unlock()
	if cb != nil {
		cb()
	}
	return nil
}

// StopServer gracefully shuts the HTTP server down.
func (s *AppState) StopServer() {
	s.mu.Lock()
	if !s.ServerRunning || s.serverStop == nil {
		s.mu.Unlock()
		return
	}
	close(s.serverStop)
	s.serverStop = nil
	s.mu.Unlock()
}

// â”€â”€â”€ HTTP routes â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€

func (s *AppState) setupRoutes(mux *http.ServeMux) {
	resp := response.New(func(i response.ResponseItem) {
		s.completeJob(i.OldID, i.NewID)
		if s.exportJSON {
			s.mu.Lock()
			s.respHistory = append(s.respHistory, i)
			j, err := json.Marshal(s.respHistory)
			s.mu.Unlock()
			if err != nil {
				log.Println("marshal:", err)
				return
			}
			if err := files.Write(s.exportName, string(j)); err != nil {
				log.Println("write export:", err)
			}
		}
	})
	s.mu.Lock()
	s.resp = resp
	s.mu.Unlock()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		s.mu.Lock()
		hasItems := resp.Len() > 0
		busy := s.Busy
		finished := s.finished
		s.mu.Unlock()

		if !hasItems && !busy {
			if !finished {
				s.mu.Lock()
				s.finished = true
				s.Busy = false
				s.exportJSON = false
				s.respHistory = make([]response.ResponseItem, 0)
				s.mu.Unlock()

				resp.Clear()
				fmt.Fprint(w, "done")
				s.AddLog("success", "Finished reuploading. (plugin can rerun without restarting)")
				s.mu.Lock()
				cb := s.OnState
				s.mu.Unlock()
				if cb != nil {
					cb()
				}
			}
			return
		}
		if err := resp.EncodeJSON(json.NewEncoder(w)); err != nil {
			log.Println("encode:", err)
		} else {
			resp.Clear()
		}
	})

	mux.HandleFunc("POST /reupload", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		busy := s.Busy
		finished := s.finished
		s.mu.Unlock()

		if busy || !finished {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		client, authErr := s.clientForStudioUpload()
		if authErr != nil {
			s.AddLog("error", authErr.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req request.RawRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.AddLog("error", "Decode request: "+err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !assets.DoesModuleExist(req.AssetType) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if client.IsOAuthOnly() && req.AssetType != "Animation" && req.AssetType != "Sound" {
			s.AddLog("error", req.AssetType+" replacement is not available with Roblox OAuth yet. Use Animation in the Lettuce Studio plugin.")
			http.Error(w, "OAuth currently supports Animation replacement only", http.StatusUnprocessableEntity)
			return
		}

		startReupload, err := assets.NewReuploadHandlerWithType(req.AssetType, client, &req, resp)
		if err != nil {
			s.AddLog("error", "Handler: "+err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.ExportJSON {
			s.mu.Lock()
			s.exportJSON = true
			s.exportName = fmt.Sprintf("Output_%s_%s.json", req.AssetType, time.Now().Format("2006-01-02_15-04-05"))
			s.mu.Unlock()
		}
		s.startJobs(req.IDs)

		s.mu.Lock()
		s.Busy = true
		s.finished = false
		cb := s.OnState
		s.mu.Unlock()
		if cb != nil {
			cb()
		}

		s.AddLog("info", fmt.Sprintf("Starting reupload of %d %s(s)â€¦", len(req.IDs), req.AssetType))

		go func() {
			start := time.Now()
			err := startReupload()

			s.mu.Lock()
			s.Busy = false
			s.mu.Unlock()
			s.finishJobs()

			if err != nil {
				s.mu.Lock()
				s.finished = true
				s.mu.Unlock()
				s.AddLog("error", "Reupload failed: "+err.Error())
				s.mu.Lock()
				cb := s.OnState
				s.mu.Unlock()
				if cb != nil {
					cb()
				}
				return
			}

			d := time.Since(start)
			s.AddLog("info", fmt.Sprintf(
				"Done in %dh %dm %ds â€” waiting for plugin to finish swapping IDsâ€¦",
				int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60,
			))
			s.mu.Lock()
			cb2 := s.OnState
			s.mu.Unlock()
			if cb2 != nil {
				cb2()
			}
		}()

		w.WriteHeader(http.StatusOK)
	})
}
