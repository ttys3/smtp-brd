#!/bin/sh

# merge master code to ctr branch to trigger container image build action

git checkout ctr && \
git merge --ff-only master && \
git push origin ctr && \
git checkout master && \
echo "all done"