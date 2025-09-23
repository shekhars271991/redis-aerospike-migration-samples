#!/bin/bash

# Setup Go PATH permanently
# This script adds Go to your shell profile so it persists across sessions

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "🔧 Go PATH Setup"
echo "================"

# Check if Go is already in PATH
if command -v go &> /dev/null; then
    print_success "Go $(go version | awk '{print $3}') is already available"
    exit 0
fi

# Check if Go exists in common locations
GO_PATH=""
if [[ -x "/usr/local/go/bin/go" ]]; then
    GO_PATH="/usr/local/go/bin"
elif [[ -x "$HOME/go/bin/go" ]]; then
    GO_PATH="$HOME/go/bin"
else
    print_error "Go not found in common locations (/usr/local/go/bin or $HOME/go/bin)"
    print_info "Please install Go first using one of the run scripts:"
    print_info "  ./run-quiz-app.sh"
    print_info "  ./run-loadtest.sh"
    exit 1
fi

print_info "Found Go at: $GO_PATH"

# Determine which shell profile to update
SHELL_PROFILE=""
if [[ -n "$ZSH_VERSION" ]]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [[ -n "$BASH_VERSION" ]]; then
    if [[ -f "$HOME/.bash_profile" ]]; then
        SHELL_PROFILE="$HOME/.bash_profile"
    else
        SHELL_PROFILE="$HOME/.bashrc"
    fi
else
    # Try to detect based on available files
    if [[ -f "$HOME/.zshrc" ]]; then
        SHELL_PROFILE="$HOME/.zshrc"
    elif [[ -f "$HOME/.bash_profile" ]]; then
        SHELL_PROFILE="$HOME/.bash_profile"
    elif [[ -f "$HOME/.bashrc" ]]; then
        SHELL_PROFILE="$HOME/.bashrc"
    fi
fi

if [[ -z "$SHELL_PROFILE" ]]; then
    print_error "Could not determine shell profile to update"
    print_info "Please manually add this line to your shell profile:"
    print_info "  export PATH=\"$GO_PATH:\$PATH\""
    exit 1
fi

print_info "Using shell profile: $SHELL_PROFILE"

# Check if PATH is already in the profile
if grep -q "$GO_PATH" "$SHELL_PROFILE" 2>/dev/null; then
    print_warning "Go PATH already exists in $SHELL_PROFILE"
else
    # Add Go to PATH
    echo "" >> "$SHELL_PROFILE"
    echo "# Go PATH (added by setup-go.sh)" >> "$SHELL_PROFILE"
    echo "export PATH=\"$GO_PATH:\$PATH\"" >> "$SHELL_PROFILE"
    print_success "Added Go PATH to $SHELL_PROFILE"
fi

# Add to current session
export PATH="$GO_PATH:$PATH"

# Verify Go is now available
if command -v go &> /dev/null; then
    print_success "Go $(go version | awk '{print $3}') is now available!"
    print_info "You may need to restart your terminal or run:"
    print_info "  source $SHELL_PROFILE"
else
    print_error "Go setup failed"
    exit 1
fi

echo
print_success "✅ Go setup complete!"
print_info "You can now run the applications without Go being downloaded each time:"
print_info "  ./run-quiz-app.sh"
print_info "  ./run-loadtest.sh"
