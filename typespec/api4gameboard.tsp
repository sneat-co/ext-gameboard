import "@typespec/http";

using Http;

/**
 * GameBoard core API — the game record and its authoritative event-timeline.
 *
 * The game lives at the GLOBAL path `/ext/gameboard/games/{gameID}` (not under
 * `/spaces`). The event log is an append-only child collection
 * `/ext/gameboard/games/{gameID}/events/{eventID}`.
 *
 * Verifies backstage feature `sports/gameboard-live/event-timeline`:
 *   - REQ:event-log, REQ:event-vocabulary, REQ:state-projection,
 *     REQ:append-only-corrections, REQ:source-authority,
 *     REQ:live-emission, REQ:record-access.
 *
 * Auth: appends require `Authorization: Bearer <token>` and the caller's
 * per-game `roles` assignment MUST authorize the event type (scorekeeper:
 * score/team-foul/substitution; timekeeper: clock/period/possession/timeout/
 * status; judge: judge-ruling/correction; or `consensus` when no crew). An
 * unauthorized append is rejected. Public read surfaces (scoreboard) need no auth.
 */
@route("/v0/api4gameboard")
namespace GameBoard.Api4GameBoard;

// ---------------------------------------------------------------------------
// Inline team side
// ---------------------------------------------------------------------------

/** Which side of the game an event/score belongs to. */
enum TeamSide {
  home,
  away,
}

/**
 * Inline descriptor stored per side on a game. `spaceID` is null for an ad-hoc
 * name (first-use-backprop fills it when a durable team space materializes) and
 * set when the side is a real sneat team space.
 */
model Side {
  name: string;
  colour: string;
  spaceID?: string | null;
}

// ---------------------------------------------------------------------------
// Event vocabulary
// ---------------------------------------------------------------------------

/** The official event vocabulary (single shared schema for all consumers). */
enum EventType {
  status,
  period,
  clock,
  score,
  `team-foul`,
  timeout,
  substitution,
  possession,
  `judge-ruling`,
  correction,
}

/** Lifecycle status of a game. */
enum GameStatus {
  scheduled,
  live,
  halftime,
  overtime,
  final,
  cancelled,
}

/** Clock event kind. */
enum ClockAction {
  start,
  stop,
  adjust,
}

/** Authorized appender of an event (per-game role or consensus). */
enum Source {
  scorekeeper,
  timekeeper,
  judge,
  consensus,
}

/**
 * One immutable entry of the log.
 *
 * `eventID` is a client-generated random dashless id used as the Firestore doc
 * key AND as the idempotency key: a replayed append with the same id is a no-op
 * (never stored twice). Total order is by `wallClockMs` ascending, ties broken
 * by `eventID`. `correctionOf` (corrections only) references the eventID this
 * event voids/amends — history is never edited in place.
 */
model Event {
  eventID: string;
  type: EventType;
  source: Source;
  wallClockMs: int64;
  period: int32;
  gameClockMs: int64;

  // payload (only fields relevant to `type` are set)
  status?: GameStatus;
  clockAction?: ClockAction;
  side?: TeamSide;
  points?: int32; // score: 1, 2 or 3
  scorerID?: string;
  assistID?: string;
  playerOn?: string;
  playerOff?: string;
  correctionOf?: string;
}

/** Body for appending an event (the client supplies eventID for idempotency). */
model AppendEventRequest {
  ...Event;
}

/** Result of an append: `applied` is false when the eventID was already processed. */
model AppendEventResponse {
  eventID: string;
  applied: boolean;
  status: string; // "appended" | "already-processed"
}

/** The deterministic projection (fold) of the log up to its end. */
model GameState {
  status: GameStatus;
  period: int32;
  gameClockMs: int64;
  clockRunning: boolean;
  scores: Record<int32>; // keyed by TeamSide
  teamFouls: Record<int32>;
  timeoutsUsed: Record<int32>;
  possession?: TeamSide;
  onCourt: Record<string[]>;
}

// ---------------------------------------------------------------------------
// Game record
// ---------------------------------------------------------------------------

/**
 * The game aggregate at `/ext/gameboard/games/{gameID}`. `scheduledMs` is the
 * start instant (epoch ms); `location` is a free-text venue. `createdBy` is the
 * authenticated organizer who created the game (the only user permitted to edit
 * its settings until crew-based authorization lands).
 */
model Game {
  gameID: string;
  home: Side;
  away: Side;
  status: GameStatus;
  scheduledMs?: int64;
  location?: string;
  createdBy?: string;
  createdAt?: string;
}

/** Body for creating a game (organizer authenticated; returns the new Game). */
model CreateGameRequest {
  home: Side;
  away: Side;
  scheduledMs?: int64;
}

/**
 * Body for editing a game's settings. Omitted fields are left unchanged. The
 * write is restricted to the game's creator (organizer); others get 403.
 */
model UpdateGameRequest {
  scheduledMs?: int64;
  location?: string;
}

// ---------------------------------------------------------------------------
// Operations
// ---------------------------------------------------------------------------

interface Games {
  /** Create a game (authenticated organizer). Returns the new game record. */
  @post
  @route("/games")
  create(@body body: CreateGameRequest): Game;

  /** Read a single game record (public, no-login). */
  @get
  @route("/games/{gameID}")
  get(@path gameID: string): Game;

  /**
   * Edit a game's schedule and/or venue. Authenticated; restricted to the
   * game's creator (organizer) — others get 403. Exposed as PUT (the shared
   * sneat API client has no PATCH); omitted fields are left unchanged.
   */
  @put
  @route("/games/{gameID}")
  update(@path gameID: string, @body body: UpdateGameRequest): Game;
}

interface Events {
  /** Append an event to a game's authoritative log (idempotent by eventID). */
  @post
  @route("/games/{gameID}/events")
  append(@path gameID: string, @body body: AppendEventRequest): AppendEventResponse;

  /** Read the complete ordered log (or as-of a wall-clock), both clocks intact. */
  @get
  @route("/games/{gameID}/events")
  list(@path gameID: string, @query asOfWallClockMs?: int64): Event[];

  /** Read the current deterministic fold (public, no-login scoreboard read). */
  @get
  @route("/games/{gameID}/state")
  state(@path gameID: string): GameState;
}
