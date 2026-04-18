#!/bin/bash
set -e

USERNAME=$1

if [ -z "$USERNAME" ]; then
    echo "Error: Username argument required." >&2
    exit 1
fi

ALLOWED_COMMANDS="/usr/bin/systemctl start, /usr/bin/systemctl start *, /usr/bin/systemctl stop, /usr/bin/systemctl stop *, /usr/bin/systemctl restart, /usr/bin/systemctl restart *, /usr/bin/systemctl status, /usr/bin/systemctl status *, /usr/bin/journalctl -f, /usr/bin/journalctl -f *, /usr/bin/journalctl -u, /usr/bin/journalctl -u *"
SUDOERS_CONTENT="$USERNAME ALL=(ALL) NOPASSWD: $ALLOWED_COMMANDS"

TEMP_FILE=$(mktemp)
FINAL_FILE="/etc/sudoers.d/vm-platform-$USERNAME"

echo "Generating sudoers configuration for $USERNAME..."
echo "$SUDOERS_CONTENT" > "$TEMP_FILE"

echo "Validating sudoers syntax..."
if visudo -c -f "$TEMP_FILE"; then
    echo "Validation successful. Applying configuration..."
    mv "$TEMP_FILE" "$FINAL_FILE"
    chmod 0440 "$FINAL_FILE"
    chown root:root "$FINAL_FILE"
    echo "Provisioning complete."
    exit 0
else
    echo "Error: Sudoers validation failed. Aborting." >&2
    rm -f "$TEMP_FILE"
    exit 1
fi