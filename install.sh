#!/bin/sh
# Installs the latest coffee release into /usr/local/bin.
# Usage: curl -fsSL https://raw.githubusercontent.com/mrthiti/coffee/main/install.sh | sh
set -eu

REPO="mrthiti/coffee"
INSTALL_DIR="/usr/local/bin"

os="$(uname -s)"
if [ "$os" != "Darwin" ]; then
	echo "Error: coffee only supports macOS (Darwin), detected: $os" >&2
	exit 1
fi

case "$(uname -m)" in
	arm64) arch="arm64" ;;
	x86_64) arch="amd64" ;;
	*)
		echo "Error: unsupported architecture: $(uname -m)" >&2
		exit 1
		;;
esac

url="https://github.com/$REPO/releases/latest/download/coffee-darwin-$arch"
tmpfile="$(mktemp)"
sudoers_tmp="$(mktemp)"
trap 'rm -f "$tmpfile" "$sudoers_tmp"' EXIT

echo "Downloading coffee ($arch) from $url ..."
curl -fsSL "$url" -o "$tmpfile"
chmod +x "$tmpfile"

# curl downloads don't carry the com.apple.quarantine attribute the way
# browser downloads do, but strip it defensively in case some environment
# tags it anyway.
xattr -dr com.apple.quarantine "$tmpfile" 2>/dev/null || true

# Prepare (but don't install yet) the sudoers rule that grants this user
# NOPASSWD access to exactly these two pmset invocations — nothing else —
# so plain "coffee" never has to prompt for a password. visudo -c -f only
# reads/parses our own temp file here, so this check needs no privilege.
user="$(id -un)"
printf '%s ALL=(root) NOPASSWD: /usr/bin/pmset -a disablesleep 0, /usr/bin/pmset -a disablesleep 1\n' "$user" >"$sudoers_tmp"
sudoers_ready=0
visudo -c -f "$sudoers_tmp" >/dev/null 2>&1 && sudoers_ready=1

# Everything below that needs root is bundled into a single sudo
# invocation so you're only asked for a password once, not once per step.
if [ -w "$INSTALL_DIR" ]; then
	install -m 755 "$tmpfile" "$INSTALL_DIR/coffee"
	if [ "$sudoers_ready" = 1 ]; then
		sudo install -m 440 "$sudoers_tmp" /etc/sudoers.d/coffee || true
	fi
elif [ "$sudoers_ready" = 1 ]; then
	sudo sh -c "install -m 755 '$tmpfile' '$INSTALL_DIR/coffee' && install -m 440 '$sudoers_tmp' /etc/sudoers.d/coffee" || true
else
	sudo install -m 755 "$tmpfile" "$INSTALL_DIR/coffee"
fi

# Best-effort: a missing binary is a hard failure, but a missing sudoers
# rule just means "coffee" will need "sudo coffee" instead — not fatal.
if [ ! -x "$INSTALL_DIR/coffee" ]; then
	echo "Error: coffee was not installed" >&2
	exit 1
fi
echo "✔ coffee installed to $INSTALL_DIR/coffee"

if [ -f /etc/sudoers.d/coffee ]; then
	echo "Run it with: coffee"
else
	echo "⚠ could not set up passwordless sudo — run 'sudo coffee' for now, or add /etc/sudoers.d/coffee yourself later (see README)" >&2
	echo "Run it with: sudo coffee"
fi
