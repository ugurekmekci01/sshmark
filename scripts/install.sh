#!/usr/bin/env bash
# sshmark install script — downloads release binary with SHA-256 verification.
# Usage: curl -fsSL <url>/install.sh | bash
set -euo pipefail

REPO="ugurekmekci01/sshmark"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
BINARY_NAME="sshmark"

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
  esac
}

main() {
  local os arch version url checksum_url tmpdir archive

  if ! command -v curl >/dev/null 2>&1; then
    echo "error: curl is required" >&2
    exit 1
  fi

  if ! command -v ssh >/dev/null 2>&1; then
    echo "warning: OpenSSH client (ssh) not found in PATH."
    echo "sshmark requires ssh. Install OpenSSH before using sshmark open."
  fi

  os="$(detect_os)"
  arch="$(detect_arch)"
  version="${SSHMARK_VERSION:-latest}"

  if [ "${version}" = "latest" ]; then
    version="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)"
  fi

  archive="${BINARY_NAME}_${version#v}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${archive}"
  checksum_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  echo "Downloading ${url}..."
  curl -fsSL -o "${tmpdir}/${archive}" "${url}"
  curl -fsSL -o "${tmpdir}/checksums.txt" "${checksum_url}"

  echo "Verifying SHA-256..."
  (
    cd "${tmpdir}"
    grep " ${archive}$" checksums.txt | shasum -a 256 -c -
  )

  tar -xzf "${tmpdir}/${archive}" -C "${tmpdir}"
  mkdir -p "${INSTALL_DIR}"
  install -m 755 "${tmpdir}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

  echo "Installed ${BINARY_NAME} ${version} to ${INSTALL_DIR}/${BINARY_NAME}"
  if ! echo ":${PATH}:" | grep -q ":${INSTALL_DIR}:"; then
    echo "Add to PATH: export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
}

main "$@"
