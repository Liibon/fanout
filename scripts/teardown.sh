#!/usr/bin/env bash
# Stop all services and remove the dataset volume.
set -euo pipefail
docker compose down -v
echo "Teardown complete — dataset volume removed."
