// Package backend is the root of the ext-gameboard backend Go module.
//
// This module holds the public contract surface of the gameboard extension —
// the event-timeline model/const shapes, the inline team-side DTO, the
// deterministic fold reducer, and the append facade interface — and depends
// only on foundational/core packages, never on another extension.
//
// TypeSpec (../typespec/api4gameboard.tsp) is the frozen wire contract; the
// Go types here are hand-implemented to match it (no emitters), per the house
// convention. The deterministic fold reducer lives in the contract so the
// backend and the frontend reducer fold the same log to identical state.
package backend
