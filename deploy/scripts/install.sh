#!/usr/bin/env bash
set -euo pipefail

# Ensure the log file exists and is writable by the wedding user.
touch /var/log/wedding-website.log
chown wedding:wedding /var/log/wedding-website.log
chmod 0640 /var/log/wedding-website.log

systemctl daemon-reload
systemctl enable wedding-website.service
