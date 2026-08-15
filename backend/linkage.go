package backend

import (
	"context"
	"errors"
	"strings"
)

// LinkedCompetition is the viewer-safe association from a GameBoard.live game
// to one contest in an external competition provider. It is deliberately
// limited to public identifiers and a public cross-link. It carries no bearer
// token, callback command/body, signer, lease, outbox state, or private
// subject.
//
// TypeSpec: ../typespec/api4gameboard.tsp (LinkedCompetition).
type LinkedCompetition struct {
	ProviderID    string `json:"providerID"`
	CompetitionID string `json:"competitionID"`
	ContestID     string `json:"contestID"`
	ContestURL    string `json:"contestURL,omitempty"`
}

// SubmissionSyncStatus is the public state of explicit result submission.
// It says nothing about how a submission is authenticated or delivered.
//
// TypeSpec: ../typespec/api4gameboard.tsp (SubmissionSyncStatus).
type SubmissionSyncStatus string

const (
	SubmissionSyncAwaitingFinal        SubmissionSyncStatus = "awaiting-final"
	SubmissionSyncReadyToSubmit        SubmissionSyncStatus = "ready-to-submit"
	SubmissionSyncSubmitting           SubmissionSyncStatus = "submitting"
	SubmissionSyncSynced               SubmissionSyncStatus = "synced"
	SubmissionSyncRequiresAdjudication SubmissionSyncStatus = "requires-adjudication"
)

// SubmissionSync is the viewer-safe result-submission projection. ResultURL,
// when present, is a permanent public result link rather than a callback or
// transport endpoint.
//
// TypeSpec: ../typespec/api4gameboard.tsp (SubmissionSync).
type SubmissionSync struct {
	Status    SubmissionSyncStatus `json:"status"`
	ResultURL string               `json:"resultURL,omitempty"`
}

// ExternalCompetitionBinding is the immutable provider/competition/contest
// identity that an external-control request is checked against. It is a
// backend port value, not a public DTO; it intentionally excludes mutable
// presentation such as a public URL.
type ExternalCompetitionBinding struct {
	ProviderID    string
	CompetitionID string
	ContestID     string
}

// ExternalControl identifies a GameBoard.live mutation whose authority is
// resolved by the external competition owner. It intentionally expresses no
// Competios policy: the provider decides which current subject may perform
// each control.
type ExternalControl string

const (
	ExternalControlScore        ExternalControl = "score"
	ExternalControlFinalize     ExternalControl = "finalize"
	ExternalControlSubmitResult ExternalControl = "submit-result"
)

// ExternalControlAuthorizationRequest is a backend-only request to resolve a
// current external scorer. CallerSubject is trusted, private ingress context;
// json:"-" prevents it from becoming a wire field and this type must never be
// used as a public projection.
type ExternalControlAuthorizationRequest struct {
	GameID        string                     `json:"-"`
	Binding       ExternalCompetitionBinding `json:"-"`
	Control       ExternalControl            `json:"-"`
	CallerSubject string                     `json:"-"`
}

var ErrInvalidExternalControlAuthorizationRequest = errors.New("gameboard contract: invalid external control authorization request")

// ValidateExternalControlAuthorizationRequest validates only the stable port
// shape. It never decides whether the caller is authorised; that remains with
// the external provider's implementation of ExternalControlAuthorizer.
func ValidateExternalControlAuthorizationRequest(request ExternalControlAuthorizationRequest) error {
	if anyBlank(
		request.GameID,
		request.Binding.ProviderID,
		request.Binding.CompetitionID,
		request.Binding.ContestID,
		request.CallerSubject,
	) || !request.Control.valid() {
		return ErrInvalidExternalControlAuthorizationRequest
	}
	return nil
}

func (control ExternalControl) valid() bool {
	switch control {
	case ExternalControlScore, ExternalControlFinalize, ExternalControlSubmitResult:
		return true
	default:
		return false
	}
}

func anyBlank(values ...string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return false
}

// ExternalControlAuthorizer resolves current external-control authority for a
// linked GameBoard.live game. Implementations authenticate and interpret the
// private CallerSubject themselves; this stable port intentionally returns no
// role, grant, token, callback, or policy details.
type ExternalControlAuthorizer interface {
	AuthorizeExternalControl(context.Context, ExternalControlAuthorizationRequest) error
}
