#!/bin/bash

# usage: $0 0.3.6 comment

version=$(echo "$1" | sed -e 's/v//g')
text=$2
branch=$(git rev-parse --abbrev-ref HEAD)
repo_full_name=$(git config --get remote.origin.url | sed 's/.*:\/\/github.com\///;s/.git$//')
token=$(git config --global github.token)

generate_post_data()
{
  cat <<EOF
{
  "tag_name": "v$version",
  "target_commitish": "$branch",
  "name": "v$version",
  "body": "$text",
  "draft": false,
  "prerelease": false
}
EOF
}

echo "bumpversion to $version"
bumpversion --new-version $version  minor
git push

echo "Create release v$version for repo: $repo_full_name branch: $branch"
curl -H "Authorization: token $token" --data "$(generate_post_data)" "https://api.github.com/repos/$repo_full_name/releases"
