#!/bin/bash

set -x

GITHUB_OUTPUT="${GITHUB_OUTPUT:=github-output.txt}"
echo '$GITHUB_OUTPUT = ' "$GITHUB_OUTPUT"

# Ensure the command is provided as an environment variable
if [ -z "$COMMAND" ]; then
  echo "Error: Please set the COMMAND environment variable."
  exit 1
fi

tmp=$(mktemp -d)
(
    cd $tmp

    # Create named pipes (FIFOs) for stdout and stderr
    mkfifo stdout.fifo stderr.fifo

    # Set up tee to capture and display stdout and stderr
    tee stdout.txt < stdout.fifo &
    tee stderr.txt < stderr.fifo >&2 &

    # Run the command, capturing stdout and stderr in the FIFOs
    ( eval "$COMMAND" ) > stdout.fifo 2> stderr.fifo

    # Wait for all background jobs to finish
    wait
)

# Expose the outputs as GitHub Actions step outputs directly from the files
# Multi-line outputs are handled with the '<<EOF' syntax
{
    echo 'stdout<<EOF'
    cat "$tmp/stdout.txt"
    echo 'EOF'
    echo 'stderr<<EOF'
    cat "$tmp/stderr.txt"
    echo 'EOF'
} > "${GITHUB_OUTPUT:=github-output.txt}"
