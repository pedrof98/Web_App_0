#!/usr/bin/env bash
echo "Removing Python cache directories..."

find . -type d -name "__pycache__" -exec rm -rf {} +

find . -type d -name ".pytest_cache" -exec rm -rf {} +

find . -type f \( -name "*.pyc" -o -name "*.pyo" -o -name "*.pyd" \) -exec rm -f {} +

echo "Cleanup complete!"

