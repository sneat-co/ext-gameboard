package backend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinkedCompetitionCompatibilityFixture(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "linked-competition", "compatibility.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		LegacyProjection  map[string]any    `json:"legacyProjection"`
		LinkedCompetition LinkedCompetition `json:"linkedCompetition"`
		SubmissionSync    SubmissionSync    `json:"submissionSync"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.LegacyProjection) != 0 {
		t.Fatalf("legacy game projection should remain valid without linkage: %#v", fixture.LegacyProjection)
	}
	if fixture.LinkedCompetition.ProviderID != "competios" || fixture.LinkedCompetition.ContestID != "semifinal-1" {
		t.Fatalf("linked competition fixture did not decode: %#v", fixture.LinkedCompetition)
	}
	if fixture.SubmissionSync.Status != SubmissionSyncReadyToSubmit {
		t.Fatalf("submission sync fixture did not decode: %#v", fixture.SubmissionSync)
	}
}

func TestPublicLinkageJSONExcludesAuthorityAndDeliveryData(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "linked-competition", "public-leak-negative.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		ForbiddenKeys []string `json:"forbiddenKeys"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}

	public := struct {
		Competition LinkedCompetition `json:"competition"`
		Sync        SubmissionSync    `json:"sync"`
	}{
		Competition: LinkedCompetition{ProviderID: "competios", CompetitionID: "summer-cup", ContestID: "semifinal-1", ContestURL: "https://competios.example/contests/semifinal-1"},
		Sync:        SubmissionSync{Status: SubmissionSyncSynced, ResultURL: "https://competios.example/results/result-1"},
	}
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range fixture.ForbiddenKeys {
		if strings.Contains(strings.ToLower(string(encoded)), strings.ToLower(forbidden)) {
			t.Fatalf("public projection leaked forbidden field %q: %s", forbidden, encoded)
		}
	}

	request, err := json.Marshal(ExternalControlAuthorizationRequest{
		GameID: "game-1",
		Binding: ExternalCompetitionBinding{
			ProviderID: public.Competition.ProviderID, CompetitionID: public.Competition.CompetitionID, ContestID: public.Competition.ContestID,
		},
		Control: ExternalControlSubmitResult, CallerSubject: "private-subject",
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(request) != "{}" {
		t.Fatalf("external control request must never serialize as a public DTO: %s", request)
	}
}

func TestValidateExternalControlAuthorizationRequest(t *testing.T) {
	valid := ExternalControlAuthorizationRequest{
		GameID:        "game-1",
		Binding:       ExternalCompetitionBinding{ProviderID: "competios", CompetitionID: "summer-cup", ContestID: "semifinal-1"},
		Control:       ExternalControlSubmitResult,
		CallerSubject: "trusted-private-subject",
	}
	if err := ValidateExternalControlAuthorizationRequest(valid); err != nil {
		t.Fatalf("valid request rejected: %v", err)
	}
	for name, mutate := range map[string]func(*ExternalControlAuthorizationRequest){
		"blank game":      func(request *ExternalControlAuthorizationRequest) { request.GameID = " " },
		"blank provider":  func(request *ExternalControlAuthorizationRequest) { request.Binding.ProviderID = "" },
		"blank subject":   func(request *ExternalControlAuthorizationRequest) { request.CallerSubject = "" },
		"unknown control": func(request *ExternalControlAuthorizationRequest) { request.Control = "grant-admin" },
	} {
		t.Run(name, func(t *testing.T) {
			request := valid
			mutate(&request)
			if err := ValidateExternalControlAuthorizationRequest(request); err != ErrInvalidExternalControlAuthorizationRequest {
				t.Fatalf("got %v, want invalid request", err)
			}
		})
	}
}
