#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;35m'
NC='\033[0m'

handle_error()
{
  local _error=$1

  echo Error ${_error} - quitting
  exit ${_error}
}

print_with_date() {
  local _message="$1"
  echo -e $(date +'%d/%m/%Y %H:%M:%S.%N') - ${_message}
}
