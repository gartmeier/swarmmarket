package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// mockRepository implements RepositoryInterface for testing.
type mockRepository struct {
	agents          map[uuid.UUID]*Agent
	agentsByHash    map[string]*Agent
	ownershipTokens map[string]*OwnershipToken
	agentsByOwner   map[uuid.UUID][]*Agent
	createErr       error
	getByIDErr      error
	updateErr       error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		agents:          make(map[uuid.UUID]*Agent),
		agentsByHash:    make(map[string]*Agent),
		ownershipTokens: make(map[string]*OwnershipToken),
		agentsByOwner:   make(map[uuid.UUID][]*Agent),
	}
}

func (m *mockRepository) Create(ctx context.Context, agent *Agent) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.agents[agent.ID] = agent
	m.agentsByHash[agent.APIKeyHash] = agent
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Agent, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if agent, ok := m.agents[id]; ok {
		return agent, nil
	}
	return nil, ErrAgentNotFound
}

func (m *mockRepository) GetByAPIKeyHash(ctx context.Context, hash string) (*Agent, error) {
	if agent, ok := m.agentsByHash[hash]; ok {
		return agent, nil
	}
	return nil, ErrAgentNotFound
}

func (m *mockRepository) Update(ctx context.Context, agent *Agent) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.agents[agent.ID]; !ok {
		return ErrAgentNotFound
	}
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	if agent, ok := m.agents[id]; ok {
		now := time.Now()
		agent.LastSeenAt = &now
		return nil
	}
	return ErrAgentNotFound
}

func (m *mockRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	if agent, ok := m.agents[id]; ok {
		agent.IsActive = false
		return nil
	}
	return ErrAgentNotFound
}

func (m *mockRepository) GetReputation(ctx context.Context, agentID uuid.UUID) (*Reputation, error) {
	agent, err := m.GetByID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return &Reputation{
		AgentID:           agent.ID,
		TrustScore:        agent.TrustScore,
		TotalTransactions: agent.TotalTransactions,
		SuccessfulTrades:  agent.SuccessfulTrades,
		AverageRating:     agent.AverageRating,
	}, nil
}

func (m *mockRepository) CreateOwnershipToken(ctx context.Context, agentID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	m.ownershipTokens[tokenHash] = &OwnershipToken{
		ID:        uuid.New(),
		AgentID:   agentID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	return nil
}

func (m *mockRepository) GetOwnershipTokenByHash(ctx context.Context, tokenHash string) (*OwnershipToken, error) {
	if token, ok := m.ownershipTokens[tokenHash]; ok {
		return token, nil
	}
	return nil, ErrTokenNotFound
}

func (m *mockRepository) MarkTokenUsed(ctx context.Context, tokenID, userID uuid.UUID) error {
	for _, token := range m.ownershipTokens {
		if token.ID == tokenID {
			now := time.Now()
			token.UsedAt = &now
			token.UsedByUserID = &userID
			return nil
		}
	}
	return ErrTokenNotFound
}

func (m *mockRepository) SetAgentOwner(ctx context.Context, agentID, userID uuid.UUID) error {
	if agent, ok := m.agents[agentID]; ok {
		agent.OwnerUserID = &userID
		agent.TrustScore = 1.0
		m.agentsByOwner[userID] = append(m.agentsByOwner[userID], agent)
		return nil
	}
	return ErrAgentNotFound
}

func (m *mockRepository) GetAgentOwnerID(ctx context.Context, agentID uuid.UUID) (*uuid.UUID, error) {
	if agent, ok := m.agents[agentID]; ok {
		return agent.OwnerUserID, nil
	}
	return nil, ErrAgentNotFound
}

func (m *mockRepository) GetAgentsByOwner(ctx context.Context, userID uuid.UUID) ([]*Agent, error) {
	return m.agentsByOwner[userID], nil
}

func (m *mockRepository) CountActiveListings(ctx context.Context, agentID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockRepository) UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error {
	return nil
}

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

// Service-level tests using mock repository

func TestService_Register(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)

	ctx := context.Background()
	req := &RegisterRequest{
		Name:        "Test Agent",
		Description: "A test agent",
		OwnerEmail:  "test@example.com",
		Metadata:    map[string]any{"version": "1.0"},
	}

	resp, err := service.Register(ctx, req)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.Agent == nil {
		t.Fatal("Agent should not be nil")
	}
	if resp.APIKey == "" {
		t.Error("APIKey should not be empty")
	}
	if !strings.HasPrefix(resp.APIKey, "sm_") {
		t.Error("APIKey should have sm_ prefix")
	}
	if resp.Agent.Name != "Test Agent" {
		t.Errorf("expected name 'Test Agent', got %s", resp.Agent.Name)
	}
	if resp.Agent.VerificationLevel != VerificationBasic {
		t.Error("new agents should have basic verification")
	}
	if resp.Agent.TrustScore != 0.5 {
		t.Error("new agents should have trust score 0.5")
	}
	if !resp.Agent.IsActive {
		t.Error("new agents should be active")
	}

	// Verify agent was stored
	stored, err := repo.GetByID(ctx, resp.Agent.ID)
	if err != nil {
		t.Fatalf("Agent not stored: %v", err)
	}
	if stored.Name != "Test Agent" {
		t.Error("stored agent has wrong name")
	}
}

func TestService_GetByID(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	// Create an agent first
	agentID := uuid.New()
	agent := &Agent{
		ID:        agentID,
		Name:      "Test Agent",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	repo.agents[agentID] = agent

	// Get by ID
	result, err := service.GetByID(ctx, agentID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if result.ID != agentID {
		t.Error("returned wrong agent")
	}

	// Get non-existent
	_, err = service.GetByID(ctx, uuid.New())
	if err != ErrAgentNotFound {
		t.Error("expected ErrAgentNotFound for non-existent agent")
	}
}

func TestService_GetPublicProfile(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{
		ID:                agentID,
		Name:              "Public Agent",
		Description:       "Description",
		OwnerEmail:        "secret@example.com",
		APIKeyHash:        "secret_hash",
		VerificationLevel: VerificationVerified,
		TrustScore:        0.9,
		IsActive:          true,
		CreatedAt:         time.Now(),
	}
	repo.agents[agentID] = agent

	profile, err := service.GetPublicProfile(ctx, agentID)
	if err != nil {
		t.Fatalf("GetPublicProfile failed: %v", err)
	}

	if profile.Name != "Public Agent" {
		t.Error("profile name incorrect")
	}
	if profile.TrustScore != 0.9 {
		t.Error("profile trust score incorrect")
	}
}

func TestService_ValidateAPIKey(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	// Register an agent to get a valid API key
	resp, err := service.Register(ctx, &RegisterRequest{
		Name:       "API Key Agent",
		OwnerEmail: "test@example.com",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Validate the API key
	agent, err := service.ValidateAPIKey(ctx, resp.APIKey)
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
	if agent.ID != resp.Agent.ID {
		t.Error("returned wrong agent")
	}

	// Validate invalid key
	_, err = service.ValidateAPIKey(ctx, "sm_invalid_key")
	if err != ErrAgentNotFound {
		t.Error("expected ErrAgentNotFound for invalid key")
	}
}

func TestService_Update(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	// Create agent
	agentID := uuid.New()
	agent := &Agent{
		ID:        agentID,
		Name:      "Original Name",
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	repo.agents[agentID] = agent

	// Update
	newName := "Updated Name"
	newDesc := "New description"
	updated, err := service.Update(ctx, agentID, &UpdateRequest{
		Name:        &newName,
		Description: &newDesc,
		Metadata:    map[string]any{"updated": true},
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updated.Name)
	}
	if updated.Description != "New description" {
		t.Error("description not updated")
	}
	if updated.Metadata["updated"] != true {
		t.Error("metadata not updated")
	}
}

func TestService_Update_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	name := "New Name"
	_, err := service.Update(ctx, uuid.New(), &UpdateRequest{Name: &name})
	if err != ErrAgentNotFound {
		t.Error("expected ErrAgentNotFound")
	}
}

func TestService_Deactivate(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{
		ID:       agentID,
		Name:     "To Deactivate",
		IsActive: true,
	}
	repo.agents[agentID] = agent

	err := service.Deactivate(ctx, agentID)
	if err != nil {
		t.Fatalf("Deactivate failed: %v", err)
	}

	// Verify deactivated
	if repo.agents[agentID].IsActive {
		t.Error("agent should be deactivated")
	}
}

func TestService_GetReputation(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{
		ID:                agentID,
		Name:              "Reputable Agent",
		TrustScore:        0.95,
		TotalTransactions: 100,
		SuccessfulTrades:  98,
		AverageRating:     4.8,
		IsActive:          true,
	}
	repo.agents[agentID] = agent

	rep, err := service.GetReputation(ctx, agentID)
	if err != nil {
		t.Fatalf("GetReputation failed: %v", err)
	}

	if rep.TrustScore != 0.95 {
		t.Errorf("expected trust score 0.95, got %f", rep.TrustScore)
	}
	if rep.TotalTransactions != 100 {
		t.Error("total transactions incorrect")
	}
	if rep.SuccessfulTrades != 98 {
		t.Error("successful trades incorrect")
	}
}

func TestService_GenerateOwnershipToken(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{ID: agentID, Name: "Token Agent", IsActive: true}
	repo.agents[agentID] = agent

	token, expiresAt, err := service.GenerateOwnershipToken(ctx, agentID)
	if err != nil {
		t.Fatalf("GenerateOwnershipToken failed: %v", err)
	}

	if !strings.HasPrefix(token, "own_") {
		t.Error("token should have own_ prefix")
	}
	if expiresAt.Before(time.Now()) {
		t.Error("expiry should be in the future")
	}

	// Verify token was stored
	if len(repo.ownershipTokens) != 1 {
		t.Error("token should be stored")
	}
}

func TestService_GenerateOwnershipToken_NotFound(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	_, _, err := service.GenerateOwnershipToken(ctx, uuid.New())
	if err != ErrAgentNotFound {
		t.Error("expected ErrAgentNotFound")
	}
}

func TestService_ClaimOwnership(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	// Create agent
	agentID := uuid.New()
	agent := &Agent{ID: agentID, Name: "Claimable Agent", IsActive: true, TrustScore: 0.5}
	repo.agents[agentID] = agent

	// Generate token
	token, _, err := service.GenerateOwnershipToken(ctx, agentID)
	if err != nil {
		t.Fatalf("GenerateOwnershipToken failed: %v", err)
	}

	// Claim ownership
	userID := uuid.New()
	claimed, err := service.ClaimOwnership(ctx, userID, token)
	if err != nil {
		t.Fatalf("ClaimOwnership failed: %v", err)
	}

	if claimed.OwnerUserID == nil || *claimed.OwnerUserID != userID {
		t.Error("owner not set correctly")
	}
	if claimed.TrustScore != 1.0 {
		t.Error("trust score should be boosted to 1.0")
	}
}

func TestService_ClaimOwnership_ExpiredToken(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{ID: agentID, Name: "Agent", IsActive: true}
	repo.agents[agentID] = agent

	// Create expired token directly
	tokenHash := service.hashToken("own_test_expired")
	repo.ownershipTokens[tokenHash] = &OwnershipToken{
		ID:        uuid.New(),
		AgentID:   agentID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-25 * time.Hour),
	}

	_, err := service.ClaimOwnership(ctx, uuid.New(), "own_test_expired")
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestService_ClaimOwnership_UsedToken(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agentID := uuid.New()
	agent := &Agent{ID: agentID, Name: "Agent", IsActive: true}
	repo.agents[agentID] = agent

	// Create already used token
	tokenHash := service.hashToken("own_test_used")
	usedAt := time.Now().Add(-1 * time.Hour)
	usedBy := uuid.New()
	repo.ownershipTokens[tokenHash] = &OwnershipToken{
		ID:           uuid.New(),
		AgentID:      agentID,
		TokenHash:    tokenHash,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		UsedAt:       &usedAt,
		UsedByUserID: &usedBy,
		CreatedAt:    time.Now().Add(-1 * time.Hour),
	}

	_, err := service.ClaimOwnership(ctx, uuid.New(), "own_test_used")
	if err != ErrTokenAlreadyUsed {
		t.Errorf("expected ErrTokenAlreadyUsed, got %v", err)
	}
}

func TestService_ClaimOwnership_AlreadyOwned(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	// Create agent with existing owner
	agentID := uuid.New()
	existingOwner := uuid.New()
	agent := &Agent{ID: agentID, Name: "Owned Agent", IsActive: true, OwnerUserID: &existingOwner}
	repo.agents[agentID] = agent

	// Create valid token
	tokenHash := service.hashToken("own_test_valid")
	repo.ownershipTokens[tokenHash] = &OwnershipToken{
		ID:        uuid.New(),
		AgentID:   agentID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err := service.ClaimOwnership(ctx, uuid.New(), "own_test_valid")
	if err != ErrAgentAlreadyOwned {
		t.Errorf("expected ErrAgentAlreadyOwned, got %v", err)
	}
}

func TestService_GetAgentsByOwner(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	userID := uuid.New()

	// Create agents owned by user
	agent1 := &Agent{ID: uuid.New(), Name: "Agent 1", OwnerUserID: &userID}
	agent2 := &Agent{ID: uuid.New(), Name: "Agent 2", OwnerUserID: &userID}
	repo.agents[agent1.ID] = agent1
	repo.agents[agent2.ID] = agent2
	repo.agentsByOwner[userID] = []*Agent{agent1, agent2}

	agents, err := service.GetAgentsByOwner(ctx, userID)
	if err != nil {
		t.Fatalf("GetAgentsByOwner failed: %v", err)
	}

	if len(agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(agents))
	}
}

func TestService_GetAgentsByOwner_Empty(t *testing.T) {
	repo := newMockRepository()
	service := NewService(repo, 32)
	ctx := context.Background()

	agents, err := service.GetAgentsByOwner(ctx, uuid.New())
	if err != nil {
		t.Fatalf("GetAgentsByOwner failed: %v", err)
	}

	if len(agents) != 0 {
		t.Error("expected empty list for user with no agents")
	}
}
