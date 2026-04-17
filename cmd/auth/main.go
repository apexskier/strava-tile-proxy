// cmd/auth captures Strava session cookies via a browser-assisted Google OAuth
// flow and saves them to a JSON session file for use by the tile proxy.
//
// Usage:
//
//	go run ./cmd/auth [-session /path/to/session.json]
//
// A Chrome window will open. Log in with Google, then wait — the tool
// detects when the login completes and saves the cookies automatically.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

func main() {
	defaultPath := ".env.auth"
	sessionPath := flag.String("session", defaultPath, "path to save session file")
	flag.Parse()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	fmt.Println("Opening Strava login — please log in with Google...")

	if err := chromedp.Run(ctx, chromedp.Navigate("https://www.strava.com/login")); err != nil {
		log.Fatalf("navigate: %v", err)
	}

	// Poll until the user completes OAuth and lands on /dashboard or /athlete/dashboard.
	fmt.Println("Waiting for login to complete (up to 10 minutes)...")
	loginCtx, loginCancel := context.WithTimeout(ctx, 10*time.Minute)
	defer loginCancel()

	for {
		var location string
		if err := chromedp.Run(loginCtx, chromedp.Location(&location)); err != nil {
			log.Fatalf("get location: %v", err)
		}
		if strings.Contains(location, "/dashboard") || strings.Contains(location, "/maps") {
			break
		}
		select {
		case <-loginCtx.Done():
			log.Fatal("timed out waiting for login")
		case <-time.After(time.Second):
		}
	}

	fmt.Println("Login detected — extracting cookies...")

	var rememberToken, stravaSession string
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		cookies, err := storage.GetCookies().Do(ctx)
		if err != nil {
			return err
		}
		for _, c := range cookies {
			switch c.Name {
			case "strava_remember_token":
				rememberToken = c.Value
			case "_strava4_session":
				stravaSession = c.Value
			}
		}
		return nil
	})); err != nil {
		log.Fatalf("get cookies: %v", err)
	}

	if rememberToken == "" {
		log.Fatal("strava_remember_token cookie not found")
	}
	if stravaSession == "" {
		log.Fatal("_strava4_session cookie not found")
	}

	content := fmt.Sprintf("STRAVA_REMEMBER_TOKEN=%s\nSTRAVA4_SESSION=%s\n", rememberToken, stravaSession)
	if err := os.MkdirAll(filepath.Dir(*sessionPath), 0700); err != nil {
		log.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(*sessionPath, []byte(content), 0600); err != nil {
		log.Fatalf("write session: %v", err)
	}

	fmt.Printf("Session saved to %s\n", *sessionPath)
}
