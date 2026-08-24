#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKS_DIR="${REPO_ROOT}/.githooks"

chmod +x "${HOOKS_DIR}/pre-commit"
git -C "${REPO_ROOT}" config core.hooksPath .githooks

echo "Installed .githooks/pre-commit (runs make ci)"
echo "core.hooksPath set to .githooks"
