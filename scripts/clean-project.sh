#!/usr/bin/env bash
#
# This script cleans the project directory by removing temporary files and build artifacts.
#
echo "==> Cleaning project..."

# Validates we have items to remove.
if [ ${#OUTPUTS[@]} -eq 0 ]; then echo "ERROR: Set OUTPUTS to the names of items to remove"; exit 1; fi

declare -a List=${OUTPUTS[*]}

# Iterates output array values and removes each, with force if necessary.
for s in "${List[@]}"
  do rm -f "${s}";
done

echo "==> Project Cleaned!"