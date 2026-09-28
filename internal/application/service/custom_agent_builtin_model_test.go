package service

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/require"
)

type builtinModelDefaultAgentRepo struct {
	interfaces.CustomAgentRepository
	agents []*types.CustomAgent
}

func (r *builtinModelDefaultAgentRepo) GetAgentByID(
	_ context.Context, id string, tenantID uint64,
) (*types.CustomAgent, error) {
	for _, agent := range r.agents {
		if agent.ID == id && agent.TenantID == tenantID {
			return agent, nil
		}
	}
	return nil, nil
}

func (r *builtinModelDefaultAgentRepo) ListAgentsByTenantID(
	_ context.Context, tenantID uint64,
) ([]*types.CustomAgent, error) {
	result := make([]*types.CustomAgent, 0, len(r.agents))
	for _, agent := range r.agents {
		if agent.TenantID == tenantID {
			result = append(result, agent)
		}
	}
	return result, nil
}

func loadBuiltinAgentsForModelDefaultTest(t *testing.T) {
	t.Helper()
	require.NoError(t, types.LoadBuiltinAgentsConfig(filepath.Join("..", "..", "..", "config")))
}

func persistedQuickAnswerWithoutModel(tenantID uint64) *types.CustomAgent {
	return &types.CustomAgent{
		ID:        types.BuiltinQuickAnswerID,
		TenantID:  tenantID,
		IsBuiltin: true,
		Config: types.CustomAgentConfig{
			AgentMode:   types.AgentModeQuickAnswer,
			ModelID:     "",
			Temperature: 0.25,
		},
	}
}

func TestGetBuiltinAgentInheritsDefaultModelForPersistedConfig(t *testing.T) {
	loadBuiltinAgentsForModelDefaultTest(t)

	const tenantID = uint64(42)
	svc := &customAgentService{repo: &builtinModelDefaultAgentRepo{
		agents: []*types.CustomAgent{persistedQuickAnswerWithoutModel(tenantID)},
	}}
	ctx := context.WithValue(context.Background(), types.TenantIDContextKey, tenantID)

	agent, err := svc.GetAgentByID(ctx, types.BuiltinQuickAnswerID)
	require.NoError(t, err)
	require.Equal(t, "builtin-llm-default", agent.Config.ModelID)
	require.Equal(t, 0.25, agent.Config.Temperature, "tenant customizations must be preserved")
}

func TestListAgentsInheritsDefaultModelForPersistedBuiltin(t *testing.T) {
	loadBuiltinAgentsForModelDefaultTest(t)

	const tenantID = uint64(42)
	svc := &customAgentService{repo: &builtinModelDefaultAgentRepo{
		agents: []*types.CustomAgent{persistedQuickAnswerWithoutModel(tenantID)},
	}}
	ctx := context.WithValue(context.Background(), types.TenantIDContextKey, tenantID)

	agents, err := svc.ListAgents(ctx)
	require.NoError(t, err)
	for _, agent := range agents {
		if agent.ID == types.BuiltinQuickAnswerID {
			require.Equal(t, "builtin-llm-default", agent.Config.ModelID)
			require.Equal(t, 0.25, agent.Config.Temperature, "tenant customizations must be preserved")
			return
		}
	}
	t.Fatal("built-in quick-answer agent was not returned")
}
