package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAPIKey(t *testing.T) {
	s := NewService(nil, 32)

	key, err := s.generateAPIKey()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check prefix
	if !strings.HasPrefix(key, "sm_") {
		t.Errorf("expected key to have prefix 'sm_', got %s", key)
	}

	// Check length: "sm_" (3) + 64 hex chars (32 bytes * 2)
	expectedLen := 3 + 64
	if len(key) != expectedLen {
		t.Errorf("expected key length %d, got %d", expectedLen, len(key))
	}
}

func TestHashAPIKey(t *testing.T) {
	s := NewService(nil, 32)

	key := "sm_test_key_12345"
	hash1 := s.hashAPIKey(key)
	hash2 := s.hashAPIKey(key)

	// Same input should produce same hash
	if hash1 != hash2 {
		t.Error("expected same hash for same input")
	}

	// Different input should produce different hash
	hash3 := s.hashAPIKey("sm_different_key")
	if hash1 == hash3 {
		t.Error("expected different hash for different input")
	}

	// Hash should be 64 chars (SHA-256 = 32 bytes = 64 hex chars)
	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestNewServiceDefaultKeyLength(t *testing.T) {
	s := NewService(nil, 0)

	if s.keyLength != 32 {
		t.Errorf("expected default key length 32, got %d", s.keyLength)
	}
}

func TestAgentPublicProfile(t *testing.T) {
	agent := &Agent{
		Name:              "Test Agent",
		Description:       "A test agent",
		OwnerEmail:        "secret@example.com",
		APIKeyHash:        "secrethash",
		VerificationLevel: VerificationBasic,
		TrustScore:        0.8,
		TotalTransactions: 10,
		SuccessfulTrades:  9,
		AverageRating:     4.5,
	}

	profile := agent.PublicProfile()

	// Should include public fields
	if profile.Name != agent.Name {
		t.Errorf("expected name %s, got %s", agent.Name, profile.Name)
	}
	if profile.TrustScore != agent.TrustScore {
		t.Errorf("expected trust score %f, got %f", agent.TrustScore, profile.TrustScore)
	}

	// Should not include sensitive fields (checked by type system)
	// OwnerEmail and APIKeyHash are not in AgentPublicProfile
}

func TestNewServiceNegativeKeyLength(t *testing.T) {
	s := NewService(nil, -5)
	if s.keyLength != 32 {
		t.Errorf("expected default key length 32 for negative input, got %d", s.keyLength)
	}
}

func TestNewServiceCustomKeyLength(t *testing.T) {
	s := NewService(nil, 64)
	if s.keyLength != 64 {
		t.Errorf("expected key length 64, got %d", s.keyLength)
	}
}

func TestGenerateAPIKeyUniqueness(t *testing.T) {
	s := NewService(nil, 32)

	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := s.generateAPIKey()
		if err != nil {
			t.Fatalf("failed to generate key: %v", err)
		}
		if keys[key] {
			t.Error("generated duplicate key")
		}
		keys[key] = true
	}
}

func TestGenerateOwnershipToken(t *testing.T) {
	s := NewService(nil, 32)

	token, err := s.generateOwnershipToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if !strings.HasPrefix(token, "own_") {
		t.Errorf("expected token to start with 'own_', got %s", token[:10])
	}

	// Check length (own_ + 48 hex chars = 52 total)
	expectedLen := 4 + 48
	if len(token) != expectedLen {
		t.Errorf("expected token length %d, got %d", expectedLen, len(token))
	}
}

func TestHashToken(t *testing.T) {
	s := NewService(nil, 32)

	hash1 := s.hashToken("own_test123")
	hash2 := s.hashToken("own_test123")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}

	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestVerificationLevel(t *testing.T) {
	tests := []struct {
		level    VerificationLevel
		expected string
	}{
		{VerificationBasic, "basic"},
		{VerificationVerified, "verified"},
		{VerificationPremium, "premium"},
	}

	for _, tt := range tests {
		if string(tt.level) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.level)
		}
	}
}

func TestRegisterRequest(t *testing.T) {
	req := &RegisterRequest{
		Name:        "Test Agent",
		Description: "A test agent",
		OwnerEmail:  "test@example.com",
		Metadata: map[string]any{
			"version": "1.0",
		},
	}

	if req.Name != "Test Agent" {
		t.Errorf("expected name 'Test Agent', got %s", req.Name)
	}
	if req.OwnerEmail != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", req.OwnerEmail)
	}
	if req.Metadata["version"] != "1.0" {
		t.Error("metadata not set correctly")
	}
}

func TestRegisterResponse(t *testing.T) {
	agent := &Agent{
		ID:   uuid.New(),
		Name: "Test Agent",
	}
	resp := &RegisterResponse{
		Agent:  agent,
		APIKey: "sm_test123",
	}

	if resp.Agent != agent {
		t.Error("Agent not set correctly")
	}
	if resp.APIKey != "sm_test123" {
		t.Errorf("expected API key 'sm_test123', got %s", resp.APIKey)
	}
}

func TestUpdateRequest(t *testing.T) {
	name := "New Name"
	desc := "New Description"

	req := &UpdateRequest{
		Name:        &name,
		Description: &desc,
		Metadata: map[string]any{
			"updated": true,
		},
	}

	if *req.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %s", *req.Name)
	}
	if *req.Description != "New Description" {
		t.Errorf("expected description 'New Description', got %s", *req.Description)
	}
}

func TestReputation(t *testing.T) {
	agentID := uuid.New()
	rep := &Reputation{
		AgentID:           agentID,
		TrustScore:        0.9,
		TotalTransactions: 50,
		SuccessfulTrades:  48,
		FailedTrades:      2,
		DisputesWon:       3,
		DisputesLost:      1,
		AverageRating:     4.8,
		RatingCount:       45,
		CategoryScores: map[string]float64{
			"data":     0.95,
			"services": 0.85,
		},
	}

	if rep.AgentID != agentID {
		t.Error("AgentID not set correctly")
	}
	if rep.TrustScore != 0.9 {
		t.Errorf("expected trust score 0.9, got %f", rep.TrustScore)
	}
	if rep.SuccessfulTrades != 48 {
		t.Errorf("expected 48 successful trades, got %d", rep.SuccessfulTrades)
	}
	if rep.CategoryScores["data"] != 0.95 {
		t.Error("category scores not set correctly")
	}
}

func TestRating(t *testing.T) {
	now := time.Now()
	rating := Rating{
		TransactionID: uuid.New(),
		RaterID:       uuid.New(),
		Score:         5,
		Comment:       "Great transaction!",
		CreatedAt:     now,
	}

	if rating.Score != 5 {
		t.Errorf("expected score 5, got %d", rating.Score)
	}
	if rating.Comment != "Great transaction!" {
		t.Errorf("expected comment 'Great transaction!', got %s", rating.Comment)
	}
}

func TestAgent_AllFields(t *testing.T) {
	now := time.Now()
	lastSeen := now.Add(-time.Hour)
	ownerID := uuid.New()

	agent := &Agent{
		ID:                uuid.New(),
		Name:              "Full Agent",
		Description:       "A fully populated agent",
		OwnerEmail:        "owner@example.com",
		OwnerUserID:       &ownerID,
		APIKeyHash:        "hash123",
		APIKeyPrefix:      "sm_abc",
		VerificationLevel: VerificationPremium,
		TrustScore:        0.95,
		TotalTransactions: 200,
		SuccessfulTrades:  195,
		AverageRating:     4.9,
		IsActive:          true,
		Metadata: map[string]any{
			"capabilities": []string{"search", "analysis"},
		},
		CreatedAt:  now,
		UpdatedAt:  now,
		LastSeenAt: &lastSeen,
	}

	if agent.OwnerUserID == nil || *agent.OwnerUserID != ownerID {
		t.Error("OwnerUserID not set correctly")
	}
	if agent.LastSeenAt == nil || !agent.LastSeenAt.Equal(lastSeen) {
		t.Error("LastSeenAt not set correctly")
	}
	if agent.Metadata == nil {
		t.Error("Metadata should not be nil")
	}
	if agent.VerificationLevel != VerificationPremium {
		t.Error("VerificationLevel not set correctly")
	}
}

func TestAgentPublicProfile_AllFields(t *testing.T) {
	now := time.Now()
	agent := &Agent{
		ID:                uuid.New(),
		Name:              "Test Agent",
		Description:       "Test description",
		OwnerEmail:        "secret@example.com",
		APIKeyHash:        "secret_hash",
		APIKeyPrefix:      "sm_abc12",
		VerificationLevel: VerificationVerified,
		TrustScore:        0.85,
		TotalTransactions: 100,
		SuccessfulTrades:  95,
		AverageRating:     4.5,
		IsActive:          true,
		CreatedAt:         now,
	}

	profile := agent.PublicProfile()

	if profile.ID != agent.ID {
		t.Error("ID not copied correctly")
	}
	if profile.Name != agent.Name {
		t.Error("Name not copied correctly")
	}
	if profile.Description != agent.Description {
		t.Error("Description not copied correctly")
	}
	if profile.VerificationLevel != agent.VerificationLevel {
		t.Error("VerificationLevel not copied correctly")
	}
	if profile.TrustScore != agent.TrustScore {
		t.Error("TrustScore not copied correctly")
	}
	if profile.TotalTransactions != agent.TotalTransactions {
		t.Error("TotalTransactions not copied correctly")
	}
	if profile.SuccessfulTrades != agent.SuccessfulTrades {
		t.Error("SuccessfulTrades not copied correctly")
	}
	if profile.AverageRating != agent.AverageRating {
		t.Error("AverageRating not copied correctly")
	}
	if !profile.CreatedAt.Equal(agent.CreatedAt) {
		t.Error("CreatedAt not copied correctly")
	}
}
