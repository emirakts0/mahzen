#!/usr/bin/env bash
# gen-certs.sh — Generate a self-signed TLS certificate for local HTTP/3 development.
#
# Produces:
#   certs/server.crt  — certificate (PEM, valid 825 days)
#   certs/server.key  — private key  (PEM, RSA 2048)
#
# The certificate includes Subject Alternative Names for localhost and 127.0.0.1
# so modern browsers and Go's TLS stack accept it without extra flags.
#
# Usage:
#   ./scripts/gen-certs.sh          # from repo root
#   make gen-certs                  # via Makefile

set -euo pipefail

CERT_DIR="$(cd "$(dirname "$0")/.." && pwd)/certs"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

mkdir -p "$CERT_DIR"

echo "Generating self-signed TLS certificate..."
echo "  Output directory : $CERT_DIR"
echo "  Certificate      : $CERT_FILE"
echo "  Private key      : $KEY_FILE"

openssl req -x509 \
  -newkey rsa:2048 \
  -keyout "$KEY_FILE" \
  -out "$CERT_FILE" \
  -days 825 \
  -nodes \
  -subj "/CN=localhost/O=Mahzen Dev/C=TR" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

echo ""
echo "Done. Update config.yaml to enable HTTP/3:"
echo ""
echo "  server:"
echo "    http:"
echo "      tls:"
echo "        cert_file: certs/server.crt"
echo "        key_file:  certs/server.key"
echo ""
echo "NOTE: Browsers will warn about this self-signed certificate."
echo "      Add certs/server.crt to your system trust store to suppress the warning."
echo "      On Linux:   sudo cp certs/server.crt /usr/local/share/ca-certificates/mahzen.crt && sudo update-ca-certificates"
echo "      On macOS:   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain certs/server.crt"
