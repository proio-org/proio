#!/bin/bash

git clone http://git@$REPO_REF repo
cd repo
git submodule init proio
git submodule update --remote proio
git add proio
git commit -m "Automatic update of proio submodule from proio Travis CI" -m "proio repository commit: $TRAVIS_COMMIT"
git push "https://$REPO_TOKEN@$REPO_REF"