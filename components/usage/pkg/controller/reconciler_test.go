// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package controller

import (
	"context"
	"github.com/gitpod-io/gitpod/usage/pkg/db"
	"github.com/gitpod-io/gitpod/usage/pkg/db/dbtest"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestUsageReconciler_ReconcileTimeRange(t *testing.T) {
	type Scenario struct {
		Name    string
		NowFunc func() time.Time

		Memberships []db.TeamMembership
		Workspaces  []db.Workspace
		Instances   []db.WorkspaceInstance

		Expected *UsageReconcileStatus
	}

	startOfMay := time.Date(2022, 05, 1, 0, 00, 00, 00, time.UTC)
	startOfJune := time.Date(2022, 06, 1, 0, 00, 00, 00, time.UTC)
	instanceStatus := []byte(`{"phase": "stopped", "conditions": {"deployed": false, "pullingImages": false, "serviceExists": false}}`)

	scenarios := []Scenario{
		(func() Scenario {
			teamID := uuid.New()
			userID := uuid.New()
			scenarioRunTime := time.Date(2022, 05, 31, 23, 00, 00, 00, time.UTC)

			membership := db.TeamMembership{
				ID:     uuid.New(),
				TeamID: teamID,
				UserID: userID,
				Role:   db.TeamMembershipRole_Member,
			}
			workspace := dbtest.NewWorkspace(t, db.Workspace{
				ID:      "gitpodio-gitpod-gyjr82jkfnd",
				OwnerID: userID,
			})
			instances := []db.WorkspaceInstance{
				// Ran throughout the reconcile period
				{
					ID:           uuid.New(),
					WorkspaceID:  workspace.ID,
					CreationTime: db.NewVarcharTime(time.Date(2022, 05, 1, 00, 00, 00, 00, time.UTC)),
					StoppedTime:  db.NewVarcharTime(time.Date(2022, 06, 1, 1, 0, 0, 0, time.UTC)),
					Status:       instanceStatus,
				},
				// Still running
				{
					ID:           uuid.New(),
					WorkspaceID:  workspace.ID,
					CreationTime: db.NewVarcharTime(time.Date(2022, 05, 30, 00, 00, 00, 00, time.UTC)),
					Status:       instanceStatus,
				},
				// No creation time, invalid record
				{
					ID:          uuid.New(),
					WorkspaceID: workspace.ID,
					StoppedTime: db.NewVarcharTime(time.Date(2022, 06, 1, 1, 0, 0, 0, time.UTC)),
					Status:      instanceStatus,
				},
			}

			expectedRuntime := instances[0].TotalRuntime(scenarioRunTime) + instances[1].TotalRuntime(scenarioRunTime)

			return Scenario{
				Name:        "oen team with one workspace",
				Memberships: []db.TeamMembership{membership},
				Workspaces:  []db.Workspace{workspace},
				Instances:   instances,
				NowFunc:     func() time.Time { return scenarioRunTime },
				Expected: &UsageReconcileStatus{
					StartTime:                 startOfMay,
					EndTime:                   startOfJune,
					WorkspaceInstances:        2,
					InvalidWorkspaceInstances: 1,
					Workspaces:                1,
					Teams:                     1,
					Report: []TeamUsage{
						{
							TeamID:            teamID.String(),
							WorkspacesRuntime: expectedRuntime,
						},
					},
				},
			}
		})(),
	}

	for _, scenario := range scenarios {
		conn := db.ConnectForTests(t)
		require.NoError(t, conn.Create(scenario.Memberships).Error)
		require.NoError(t, conn.Create(scenario.Workspaces).Error)
		require.NoError(t, conn.Create(scenario.Instances).Error)

		reconciler := &UsageReconciler{
			nowFunc: scenario.NowFunc,
			conn:    conn,
		}
		status, err := reconciler.ReconcileTimeRange(context.Background(), startOfMay, startOfJune)
		require.NoError(t, err)

		require.Equal(t, scenario.Expected, status)
	}
}