// Package apns provides an Apple Push Notification service (APNs) implementation
// of port.Notifier.
//
// This file documents the requirements and setup steps for integrating with APNs.
// It serves as a TODO guide for the developer who will implement this in the future.
//
// ============================================================
// APNs Integration Guide
// ============================================================
//
// Requirements:
//   - An APNs signing key (.p8 file) from the Apple Developer Portal
//   - Key ID (10-character string from Apple Developer)
//   - Team ID (10-character string from Apple Developer)
//   - Bundle ID of your iOS application (e.g., "com.example.app")
//
// Recommended library: github.com/sideshow/apns2
//
// Configuration to add to config struct:
//
//	type PushConfig struct {
//	    APNs APNsConfig `mapstructure:"apns"`
//	}
//
//	type APNsConfig struct {
//	    KeyPath     string `mapstructure:"key_path"`     // Path to .p8 file
//	    KeyID       string `mapstructure:"key_id"`       // 10-char key ID
//	    TeamID      string `mapstructure:"team_id"`      // 10-char team ID
//	    BundleID    string `mapstructure:"bundle_id"`    // iOS bundle ID
//	    Environment string `mapstructure:"environment"`  // "sandbox" or "production"
//	}
//
// Environment detection:
//   - Development builds → use sandbox gateway (api.development.push.apple.com:443)
//   - TestFlight/App Store → use production gateway (api.push.apple.com:443)
//
// Implementation notes:
//   - Create one notification per device token
//   - Handle response: if APNs returns InvalidToken or Unregistered, remove token from DB
//   - Handle network errors with retry + backoff
//   - Support alert, badge, sound, and custom payload fields
//   - Respect rate limits (APNs recommends no more than 1500 notifications/second/gateway)
//
// For Android (FCM), follow the same port.Notifier interface with a separate implementation.
package apns
