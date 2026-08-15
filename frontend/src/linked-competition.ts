/**
 * Viewer-safe association from a GameBoard.live game to an external
 * competition contest. It contains no authority or delivery transport data.
 *
 * TypeSpec: ../../typespec/api4gameboard.tsp (LinkedCompetition).
 */
export interface LinkedCompetition {
  providerID: string;
  competitionID: string;
  contestID: string;
  contestURL?: string;
}

/** Public status of explicit result submission; never an outbox state. */
export type SubmissionSyncStatus =
  | 'awaiting-final'
  | 'ready-to-submit'
  | 'submitting'
  | 'synced'
  | 'requires-adjudication';

/**
 * Viewer-safe result-submission projection. `resultURL`, when present, is a
 * permanent public result link rather than a callback or transport endpoint.
 *
 * TypeSpec: ../../typespec/api4gameboard.tsp (SubmissionSync).
 */
export interface SubmissionSync {
  status: SubmissionSyncStatus;
  resultURL?: string;
}
