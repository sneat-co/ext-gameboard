import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import type { LinkedCompetition, SubmissionSync } from './linked-competition.js';

interface CompatibilityFixture {
  legacyProjection: Record<string, never>;
  linkedCompetition: LinkedCompetition;
  submissionSync: SubmissionSync;
}

const compatibilityUrl = new URL('../../fixtures/linked-competition/compatibility.json', import.meta.url);
const leaksUrl = new URL('../../fixtures/linked-competition/public-leak-negative.json', import.meta.url);

describe('linked competition public contract', () => {
  it('accepts the additive compatibility fixture while an unlinked legacy projection stays valid', () => {
    const fixture: CompatibilityFixture = JSON.parse(readFileSync(fileURLToPath(compatibilityUrl), 'utf8'));
    expect(fixture.legacyProjection).toEqual({});
    expect(fixture.linkedCompetition).toMatchObject({
      providerID: 'competios', competitionID: 'summer-cup', contestID: 'semifinal-1',
    });
    expect(fixture.submissionSync.status).toBe('ready-to-submit');
  });

  it('keeps authority and delivery fields out of the public JSON projection', () => {
    const { forbiddenKeys }: { forbiddenKeys: string[] } = JSON.parse(readFileSync(fileURLToPath(leaksUrl), 'utf8'));
    const publicProjection = JSON.stringify({
      competition: {
        providerID: 'competios', competitionID: 'summer-cup', contestID: 'semifinal-1',
        contestURL: 'https://competios.example/contests/semifinal-1',
      } satisfies LinkedCompetition,
      sync: { status: 'synced', resultURL: 'https://competios.example/results/result-1' } satisfies SubmissionSync,
    });
    for (const forbidden of forbiddenKeys) {
      expect(publicProjection.toLowerCase()).not.toContain(forbidden.toLowerCase());
    }
  });
});
