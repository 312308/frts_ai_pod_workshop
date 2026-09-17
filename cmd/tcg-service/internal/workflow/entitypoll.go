// Package workflow implements the reusable EntityPollWorkflow template:
// poll a FOP entity on schedule, managing delta queries, pagination,
// validation, upsert, and watermark advancement as one durable Temporal
// workflow.
//
// Source: BRD FR-022, HLD §13 Durable Workflows ("Build reusable
// EntityPollWorkflow (Temporal-based durable workflow) template... Example:
// PollWholesaleProductsWorkflow").
package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// Activity names. Concrete implementations (FOP client, conformance
// validator, canonical repository, watermark store) are registered on the
// Temporal worker under these names; this package deliberately does not
// import those packages, so the workflow's determinism can be tested in
// isolation from I/O (AC-003f).
const (
	ActivityFetchDeltaPages  = "FetchDeltaPages"
	ActivityProcessPage      = "ProcessPage"
	ActivityAdvanceWatermark = "AdvanceWatermark"
)

// EntityPollInput is the typed input for EntityPollWorkflow (AC-003a).
type EntityPollInput struct {
	Entity string
}

// retryPolicy is the shared activity retry policy: exponential backoff on
// transient errors (AC-003c). Circuit-breaking on repeated 4xx/5xx lives in
// the activity implementations themselves (SP-20), which classify errors
// via fopclient.ResponseError before returning them to the workflow.
var retryPolicy = &temporal.RetryPolicy{
	InitialInterval:    time.Second,
	BackoffCoefficient: 2.0,
	MaximumInterval:    32 * time.Second,
	MaximumAttempts:    5,
}

var activityOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 5 * time.Minute,
	RetryPolicy:         retryPolicy,
}

// EntityPollWorkflow is the reusable template (AC-003a): fetch (delta +
// pagination) → process each page → advance watermark. ProcessPage
// internally validates every record in the page (quarantining bad ones per
// SP-02, without that alone failing the activity) and upserts the good
// ones (SP-08) plus their lineage (SP-04); it returns an error only for a
// genuine infrastructure failure (e.g. DB unreachable), not a per-record
// validation failure. Any ProcessPage error fails the workflow and — per
// BR-005 — AdvanceWatermark is never reached, so the watermark stays
// unchanged for replay (AC-003b).
func EntityPollWorkflow(ctx workflow.Context, input EntityPollInput) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var pages [][]byte
	if err := workflow.ExecuteActivity(ctx, ActivityFetchDeltaPages, input.Entity).Get(ctx, &pages); err != nil {
		return err
	}

	for _, page := range pages {
		if err := workflow.ExecuteActivity(ctx, ActivityProcessPage, input.Entity, page).Get(ctx, nil); err != nil {
			return err // BR-005: do not advance watermark on failure
		}
	}

	return workflow.ExecuteActivity(ctx, ActivityAdvanceWatermark, input.Entity).Get(ctx, nil)
}
