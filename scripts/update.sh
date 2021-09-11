#!/bin/bash
set -euo pipefail

if [ $# -lt 2 ]; then
    echo "usage: $(basename $0) <function-name> <repo-name> [tag]"
    exit 2
fi

ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
REGION=${AWS_REGION:-${AWS_DEFAULT_REGION}}

if [ -z "$REGION" ]; then
    echo Missing aws region
    exit 2
fi

FUNCTION=$1
REPO="${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/$2"
TAG=${3:-"latest"}

aws ecr get-login-password | docker login --username AWS --password-stdin $REPO

docker buildx build . -t $REPO:$TAG
docker tag $REPO:$TAG mass-exec:$TAG
docker push $REPO:$TAG

HASH=$(docker inspect $REPO:$TAG --format='{{index .RepoDigests 0}}')

aws lambda update-function-code --function-name=$FUNCTION --image-uri=$HASH | jq "{ CodeSha256: .CodeSha256 }"
echo 'Waiting for update...'
aws lambda wait function-updated --function-name=$FUNCTION && echo 'Done!' || echo 'FAILED'
