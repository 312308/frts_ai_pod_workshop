package workflow

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// registerActivityStubs binds a concrete (but bodiless — never actually
// invoked, since every test overrides behavior via env.OnActivity) function
// signature to each activity name. Temporal's TestWorkflowEnvironment
// requires a registered function per name before OnActivity can mock it;
// the real implementations are registered separately on the production
// worker (main.go), keeping this workflow package decoupled from FOP
// client/DB/validator concrete types (AC-003f: determinism-testable in
// isolation from I/O).
func registerActivityStubs(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterActivityWithOptions(
		func(ctx context.Context, entity string) ([][]byte, error) { return nil, nil },
		activity.RegisterOptions{Name: ActivityFetchDeltaPages},
	)
	env.RegisterActivityWithOptions(
		func(ctx context.Context, entity string, page []byte) error { return nil },
		activity.RegisterOptions{Name: ActivityProcessPage},
	)
	env.RegisterActivityWithOptions(
		func(ctx context.Context, entity string) error { return nil },
		activity.RegisterOptions{Name: ActivityAdvanceWatermark},
	)
}

// TestEntityPollWorkflow_HappyPath_TC_185 — AC-003b: fetch → process each
// page → advance watermark, in order, with no errors.
func TestEntityPollWorkflow_HappyPath_TC_185(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerActivityStubs(env)

	pages := [][]byte{[]byte(`{"page":1}`), []byte(`{"page":2}`)}
	env.OnActivity(ActivityFetchDeltaPages, mock.Anything, "wholesale_product").Return(pages, nil)
	env.OnActivity(ActivityProcessPage, mock.Anything, "wholesale_product", pages[0]).Return(nil)
	env.OnActivity(ActivityProcessPage, mock.Anything, "wholesale_product", pages[1]).Return(nil)
	env.OnActivity(ActivityAdvanceWatermark, mock.Anything, "wholesale_product").Return(nil)

	env.ExecuteWorkflow(EntityPollWorkflow, EntityPollInput{Entity: "wholesale_product"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

// TestEntityPollWorkflow_UpsertFailureSkipsWatermark_TC_186 — BR-005/AC-003b:
// if ProcessPage fails, AdvanceWatermark must never be called.
func TestEntityPollWorkflow_UpsertFailureSkipsWatermark_TC_186(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerActivityStubs(env)

	pages := [][]byte{[]byte(`{"page":1}`)}
	env.OnActivity(ActivityFetchDeltaPages, mock.Anything, "wholesale_product").Return(pages, nil)
	env.OnActivity(ActivityProcessPage, mock.Anything, "wholesale_product", pages[0]).
		Return(errors.New("db unreachable"))
	// Deliberately no env.OnActivity(ActivityAdvanceWatermark, ...) — if the
	// workflow calls it, the test environment fails with "no mock found".

	env.ExecuteWorkflow(EntityPollWorkflow, EntityPollInput{Entity: "wholesale_product"})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

// TestEntityPollWorkflow_FetchFailurePropagates_TC_187 — a fetch failure
// fails the workflow before any page processing or watermark advance.
func TestEntityPollWorkflow_FetchFailurePropagates_TC_187(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerActivityStubs(env)

	env.OnActivity(ActivityFetchDeltaPages, mock.Anything, "wholesale_product").
		Return(nil, errors.New("fop unreachable"))

	env.ExecuteWorkflow(EntityPollWorkflow, EntityPollInput{Entity: "wholesale_product"})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

// TestEntityPollWorkflow_NoPages_TC_188 — an empty page set (e.g. immediate
// 204) still advances the watermark — "nothing changed" is still a
// successful poll.
func TestEntityPollWorkflow_NoPages_TC_188(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	registerActivityStubs(env)

	env.OnActivity(ActivityFetchDeltaPages, mock.Anything, "wholesale_product").Return([][]byte{}, nil)
	env.OnActivity(ActivityAdvanceWatermark, mock.Anything, "wholesale_product").Return(nil)

	env.ExecuteWorkflow(EntityPollWorkflow, EntityPollInput{Entity: "wholesale_product"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}
