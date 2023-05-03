#!/usr/bin/env bash

if [ "$1" == "-h" ]; then #https://stackoverflow.com/a/30830291
  echo "Usage: $(basename "$0") [-b/--build]"
  exit 0
fi

if [ "$1" != "-h" ]; then
  echo "==== BUILD CODELAB ===="
  if ! command -v claat &> /dev/null
  then
    echo "claat required but it's not installed. Skipping."
    echo "More information: https://github.com/googlecodelabs/tools/blob/main/claat/README.md"
    # go install github.com/googlecodelabs/tools/claat@latest
    exit
  else
    # claat export *
    claat export devoxxfr2023.md
  fi
fi

echo "====  RUN CODELAB  ===="
#https://developer.mozilla.org/en-US/docs/Learn/Common_questions/Tools_and_setup/set_up_a_local_testing_server
#https://stackoverflow.com/questions/38485373/from-the-terminal-verify-if-python-3-is-installed
if command -v python3 >/dev/null 2>&1
then
  python3 -m http.server 8000
else
  python -m SimpleHTTPServer 8000
fi
