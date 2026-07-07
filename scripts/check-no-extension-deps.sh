#!/usr/bin/env bash
# Enforces the extension-contract-repo zero-other-extension-deps invariant:
# ext-gameboard must depend only on foundational/core code, never another extension.
#
# MVP mechanism (per the convention): language-native dependency-list assertions —
# `go list` for the backend module and a package.json scan for the frontend lib.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
violations=0

echo "== backend: scanning Go module dependency graph =="
if [ -f "${repo_root}/backend/go.mod" ]; then
  pushd "${repo_root}/backend" >/dev/null
  deps="$(go list -deps ./... 2>/dev/null || true)"
  bad="$(printf '%s\n' "$deps" \
    | grep -E '^github\.com/sneat-co/[a-z0-9-]+/backend' \
    | grep -v '^github\.com/sneat-co/ext-gameboard/backend' || true)"
  if [ -n "$bad" ]; then
    echo "  FORBIDDEN backend dependency on another extension:"
    printf '    %s\n' $bad
    violations=$((violations + 1))
  else
    echo "  ok — no other-extension backend dependency"
  fi
  popd >/dev/null
else
  echo "  (no backend/go.mod yet — skipped)"
fi

echo "== frontend: scanning package.json deps for @sneat/extension-* =="
fe_pkgs="$(find "${repo_root}/frontend" -name package.json -not -path '*/node_modules/*' 2>/dev/null || true)"
if [ -n "$fe_pkgs" ]; then
  for pkg in $fe_pkgs; do
    bad="$(grep -oE '@sneat/extension-[a-z0-9-]+' "$pkg" \
      | grep -v '@sneat/extension-gameboard-contract' || true)"
    if [ -n "$bad" ]; then
      echo "  FORBIDDEN frontend dependency on another extension in ${pkg#$repo_root/}:"
      printf '    %s\n' $bad
      violations=$((violations + 1))
    fi
  done
  [ "$violations" -eq 0 ] && echo "  ok — no @sneat/extension-* (other than own contract) dependency"
else
  echo "  (no frontend package.json yet — skipped)"
fi

if [ "$violations" -ne 0 ]; then
  echo "INVARIANT VIOLATED: ext-gameboard must not depend on another extension." >&2
  exit 1
fi
echo "zero-other-extension-deps invariant holds."
