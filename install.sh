#!/usr/bin/env bash
# ==========================================================
# dbspin — One-Line Installer
# https://github.com/Harshidpatel12/dbspin
# ==========================================================
set -e

REPO="Harshidpatel12/dbspin"
BINARY_NAME="dbspin"
INSTALL_DIR="/usr/local/bin"

# ----------------------------------------------------------
# Color Helpers
# ----------------------------------------------------------
BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_step()    { echo -e "\n${BOLD}${CYAN}==> ${1}${NC}"; }
log_success() { echo -e "  ${GREEN}✅${NC} $1"; }
log_error()   { echo -e "  ${RED}❌ Error:${NC} $1" >&2; exit 1; }
log_info()    { echo -e "  --> $1"; }
log_warn()    { echo -e "  ${YELLOW}⚠️ ${NC} $1"; }

# ----------------------------------------------------------
# Detect OS
# ----------------------------------------------------------
log_step "Detecting Operating System..."
OS="$(uname -s)"
case "${OS}" in
    Linux*)   OS_NAME="linux" ;;
    Darwin*)  OS_NAME="darwin" ;;
    *)        log_error "Unsupported operating system: ${OS}. Only Linux and macOS are supported." ;;
esac
log_info "Detected OS: ${OS_NAME}"

# ----------------------------------------------------------
# Detect Architecture
# ----------------------------------------------------------
log_step "Detecting Architecture..."
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)          ARCH_NAME="amd64" ;;
    aarch64 | arm64) ARCH_NAME="arm64" ;;
    *)               log_error "Unsupported CPU architecture: ${ARCH}. Only amd64 and arm64 are supported." ;;
esac
log_info "Detected Architecture: ${ARCH_NAME}"

# ----------------------------------------------------------
# Fetch Latest Release Version from GitHub API
# ----------------------------------------------------------
log_step "Fetching latest release version..."
if command -v curl &>/dev/null; then
    LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
elif command -v wget &>/dev/null; then
    LATEST_VERSION=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
else
    log_error "Neither 'curl' nor 'wget' is installed. Please install one of them and retry."
fi

if [ -z "${LATEST_VERSION}" ]; then
    log_error "Could not determine the latest release version. Make sure a release has been published at: https://github.com/${REPO}/releases"
fi
log_info "Latest version: ${LATEST_VERSION}"

# ----------------------------------------------------------
# Construct Download URL
# ----------------------------------------------------------
ARCHIVE_NAME="${BINARY_NAME}_${OS_NAME}_${ARCH_NAME}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}"
log_info "Download URL: ${DOWNLOAD_URL}"

# ----------------------------------------------------------
# Download and Install
# ----------------------------------------------------------
log_step "Downloading ${BINARY_NAME} ${LATEST_VERSION}..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

if command -v curl &>/dev/null; then
    curl -fsSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${ARCHIVE_NAME}"
else
    wget -qO "${TMP_DIR}/${ARCHIVE_NAME}" "${DOWNLOAD_URL}"
fi

log_step "Extracting archive..."
tar -xzf "${TMP_DIR}/${ARCHIVE_NAME}" -C "${TMP_DIR}"

# ----------------------------------------------------------
# Move Binary to Install Directory
# ----------------------------------------------------------
log_step "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
if [ -w "${INSTALL_DIR}" ]; then
    mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
else
    log_warn "${INSTALL_DIR} requires sudo access. Prompting for password..."
    sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
fi

# ----------------------------------------------------------
# Verify Install
# ----------------------------------------------------------
log_step "Verifying installation..."
if command -v "${BINARY_NAME}" &>/dev/null; then
    log_success "${BINARY_NAME} installed successfully!"
    echo -e "${BOLD}Run: ${CYAN}dbspin up postgres${NC}"
    echo -e "${BOLD}Example: ${CYAN}dbspin list${NC}"
    echo ""
else
    log_error "Installation failed. '${BINARY_NAME}' not found in PATH after install. Please check ${INSTALL_DIR} is in your \$PATH."
fi
