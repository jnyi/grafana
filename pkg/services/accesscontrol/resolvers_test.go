package accesscontrol_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/accesscontrol"
	"github.com/grafana/grafana/pkg/services/datasources"
)

func TestResolveKeywordScope(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.SignedInUser
		permission accesscontrol.Permission
		want       accesscontrol.Permission
		wantErr    bool
	}{
		{
			name:       "no scope",
			user:       testUser,
			permission: accesscontrol.Permission{Action: "users:read"},
			want:       accesscontrol.Permission{Action: "users:read"},
			wantErr:    false,
		},
		{
			name:       "user if resolution",
			user:       testUser,
			permission: accesscontrol.Permission{Action: "users:read", Scope: "users:self"},
			want:       accesscontrol.Permission{Action: "users:read", Scope: "users:id:2"},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		var err error
		t.Run(tt.name, func(t *testing.T) {
			resolvers := accesscontrol.NewScopeResolvers()
			scopeModifier := resolvers.GetScopeKeywordMutator(tt.user)
			tt.permission.Scope, err = scopeModifier(context.TODO(), tt.permission.Scope)
			if tt.wantErr {
				assert.Error(t, err, "expected an error during the resolution of the scope")
				return
			}
			assert.NoError(t, err)
			assert.EqualValues(t, tt.want, tt.permission, "permission did not match expected resolution")
		})
	}
}

var testUser = &models.SignedInUser{
	UserId:  2,
	OrgId:   3,
	OrgName: "TestOrg",
	OrgRole: models.ROLE_VIEWER,
	Login:   "testUser",
	Name:    "Test User",
	Email:   "testuser@example.org",
}

func TestResolveAttributeScope(t *testing.T) {
	// Calls allow us to see how many times the fakeDataSourceResolution has been called
	calls := 0
	fakeDataSourceResolver := accesscontrol.ScopeAttributeResolverFunc(func(ctx context.Context, orgID int64, initialScope string) ([]string, error) {
		calls++
		if initialScope == "datasources:name:testds" {
			return []string{accesscontrol.Scope("datasources", "id", "1")}, nil
		} else if initialScope == "datasources:name:testds2" {
			return []string{accesscontrol.Scope("datasources", "id", "2")}, nil
		} else if initialScope == "datasources:name:test:ds4" {
			return []string{accesscontrol.Scope("datasources", "id", "4")}, nil
		} else if initialScope == "datasources:name:testds5*" {
			return []string{accesscontrol.Scope("datasources", "id", "5")}, nil
		} else {
			return nil, datasources.ErrDataSourceNotFound
		}
	})

	tests := []struct {
		name          string
		orgID         int64
		evaluator     accesscontrol.Evaluator
		wantEvaluator accesscontrol.Evaluator
		wantCalls     int
		wantErr       error
	}{
		{
			name:          "should work with scope less permissions",
			evaluator:     accesscontrol.EvalPermission("datasources:read"),
			wantEvaluator: accesscontrol.EvalPermission("datasources:read"),
			wantCalls:     0,
		},
		{
			name:      "should handle an error",
			orgID:     1,
			evaluator: accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds3")),
			wantErr:   datasources.ErrDataSourceNotFound,
			wantCalls: 1,
		},
		{
			name:          "should resolve a scope",
			orgID:         1,
			evaluator:     accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds")),
			wantEvaluator: accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "1")),
			wantCalls:     1,
		},
		{
			name:  "should resolve nested scopes with cache",
			orgID: 1,
			evaluator: accesscontrol.EvalAll(
				accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds")),
				accesscontrol.EvalAny(
					accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds")),
					accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds2")),
				),
			),
			wantEvaluator: accesscontrol.EvalAll(
				accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "1")),
				accesscontrol.EvalAny(
					accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "1")),
					accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "2")),
				),
			),
			wantCalls: 2,
		},
		{
			name:          "should resolve name with colon",
			orgID:         1,
			evaluator:     accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "test:ds4")),
			wantEvaluator: accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "4")),
			wantCalls:     1,
		},
		{
			name:          "should resolve names with '*'",
			orgID:         1,
			evaluator:     accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "name", "testds5*")),
			wantEvaluator: accesscontrol.EvalPermission("datasources:read", accesscontrol.Scope("datasources", "id", "5")),
			wantCalls:     1,
		},
		{
			name:          "should return error if no resolver is found for scope",
			orgID:         1,
			evaluator:     accesscontrol.EvalPermission("dashboards:read", "dashboards:id:1"),
			wantEvaluator: nil,
			wantCalls:     0,
			wantErr:       accesscontrol.ErrResolverNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolvers := accesscontrol.NewScopeResolvers()

			// Reset calls counter
			calls = 0
			// Register a resolution method
			resolvers.AddScopeAttributeResolver("datasources:name:", fakeDataSourceResolver)

			// Test
			mutate := resolvers.GetScopeAttributeMutator(tt.orgID)
			resolvedEvaluator, err := tt.evaluator.MutateScopes(context.Background(), mutate)
			if tt.wantErr != nil {
				assert.ErrorAs(t, err, &tt.wantErr, "expected an error during the resolution of the scope")
				return
			}
			assert.NoError(t, err)
			assert.EqualValues(t, tt.wantEvaluator, resolvedEvaluator, "permission did not match expected resolution")
			assert.Equal(t, tt.wantCalls, calls, "cache has not been used")
		})
	}
}
