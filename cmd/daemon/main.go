package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	appconfig "github.com/svvly2/lettuce/internal/app/config"
	appcontext "github.com/svvly2/lettuce/internal/app/context"
	"github.com/svvly2/lettuce/internal/console"
)

type Command struct {
	Action     string `json:"action"`
	Cookie     string `json:"cookie,omitempty"`
	APIKey     string `json:"apiKey,omitempty"`
	Port       string `json:"port,omitempty"`
	CookieFile string `json:"cookieFile,omitempty"`
}

type Message struct {
	Type          string `json:"type"`
	Level         string `json:"level,omitempty"`
	Message       string `json:"message,omitempty"`
	Time          string `json:"time,omitempty"`
	IsLoggedIn    bool   `json:"isLoggedIn,omitempty"`
	Username      string `json:"username,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	ServerRunning bool   `json:"serverRunning,omitempty"`
	ServerPort    string `json:"serverPort,omitempty"`
	Busy          bool   `json:"busy,omitempty"`
}

func sendLog(level, message string) {
	msg := Message{
		Type:    "log",
		Level:   level,
		Message: message,
		Time:    time.Now().Format("15:04:05"),
	}
	out, _ := json.Marshal(msg)
	fmt.Println(string(out))
}

func sendState(state *AppState) {
	s := state.Snapshot()
	msg := Message{
		Type:          "state",
		IsLoggedIn:    s.IsLoggedIn,
		Username:      s.Username,
		DisplayName:   s.DisplayName,
		ServerRunning: s.ServerRunning,
		ServerPort:    s.ServerPort,
		Busy:          s.Busy,
	}
	out, _ := json.Marshal(msg)
	fmt.Println(string(out))
}

func sendError(err error) {
	msg := Message{
		Type:    "error",
		Message: err.Error(),
	}
	out, _ := json.Marshal(msg)
	fmt.Println(string(out))
}

func handleCommand(state *AppState, cmd Command) {
	switch cmd.Action {
	case "login_oauth":
		go func() {
			if err := state.StartOAuthLogin(); err != nil {
				sendError(fmt.Errorf("oauth login failed: %v", err))
			}
		}()
	case "logout":
		state.Logout()
	case "clear_completed":
		state.mu.Lock()
		kept := state.Queue[:0]
		for _, job := range state.Queue {
			if job.Status != "complete" {
				kept = append(kept, job)
			}
		}
		state.Queue = kept
		state.mu.Unlock()
	case "start_server":
		if err := state.StartServer(); err != nil {
			sendError(err)
		}
	case "stop_server":
		state.StopServer()
	case "update_settings":
		if cmd.Port != "" {
			appconfig.Set("port", cmd.Port)
			state.mu.Lock()
			state.ServerPort = cmd.Port
			state.mu.Unlock()
		}
		if cmd.CookieFile != "" {
			appconfig.Set("cookie_file", cmd.CookieFile)
		}
		if cmd.APIKey != "" {
			if err := state.SetAPIKey(cmd.APIKey); err != nil {
				sendError(err)
			}
		}
		if err := appconfig.Save(); err != nil {
			sendError(err)
		} else {
			sendLog("success", "Settings updated.")
			sendState(state)
		}
	default:
		sendError(fmt.Errorf("unknown action: %s", cmd.Action))
	}
}

func main() {
	state := newAppState()

	// Redirect logger and set up callbacks
	appcontext.OnLog = func(level, msg string) {
		state.AddLog(level, msg)
	}

	state.OnState = func() {
		sendState(state)
	}

	// Disable standard console features (we are headless)
	console.OnClearScreen = func() {}
	console.OnInput = func(m string) (string, error) {
		sendLog("warn", "Console input not supported in headless mode")
		return "", fmt.Errorf("headless mode")
	}
	console.OnLongInput = func(m string) (string, error) {
		sendLog("warn", "Session expired. Awaiting new cookie via UI...")
		cookie := <-state.cookieInputChan
		return cookie, nil
	}

	// Restore session at startup
	state.tryRestoreSession()
	if state.Snapshot().IsLoggedIn {
		_ = state.StartServer()
	}
	sendState(state)
	sendLog("info", "Daemon started successfully.")

	uiPort := appconfig.Get("ui_port")
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/oauth/start", func(w http.ResponseWriter, r *http.Request) {
			if state.OAuthManager == nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			url, err := state.OAuthManager.PrepareLogin()
			if err != nil {
				sendError(fmt.Errorf("oauth prepare failed: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"url": url})
		})
		mux.HandleFunc("GET /api/oauth/url", func(w http.ResponseWriter, r *http.Request) {
			if state.OAuthManager == nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			url, err := state.OAuthManager.PrepareLogin()
			if err != nil {
				sendError(fmt.Errorf("oauth prepare failed: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"url": url})
		})
		mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(state.Snapshot())
		})
		mux.HandleFunc("GET /api/logs", func(w http.ResponseWriter, r *http.Request) {
			since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(state.LogsSince(since))
		})
		mux.HandleFunc("POST /api/command", func(w http.ResponseWriter, r *http.Request) {
			var cmd Command
			if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			handleCommand(state, cmd)
			w.WriteHeader(http.StatusOK)
		})
		mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		_ = http.ListenAndServe("127.0.0.1:"+uiPort, withCORS(mux))
	}()

	// Listen for commands on stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			sendError(fmt.Errorf("failed to parse command: %v", err))
			continue
		}
		handleCommand(state, cmd)
	}

	// Electron intentionally has no interactive stdin. HTTP owns the lifecycle.
	select {}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
