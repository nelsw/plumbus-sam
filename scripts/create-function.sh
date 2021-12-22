#!/usr/bin/env bash
#
# This script creates an AWS Lambda Function.
printf "\n==> creating an AWS λƒ...\n"
if [ -z "${DOMAIN}" ]; then echo "ERROR: Set DOMAIN to the entity identity of the λƒ to build"; exit 1; fi
if [ -z "${ROLE}" ]; then echo "ERROR: Set ROLE to the AWS IAM Role of the λƒ to build"; exit 1; fi
if [ -z "${TIMEOUT}" ]; then TIMEOUT="30"; fi
if [ -z "${MEMORY}" ]; then MEMORY="512"; fi
if [ -z "${DESC}" ]; then DESC="null"; fi
if [ -f test/"${DOMAIN}"/env.json ]; then ENV="$(jq '.' test/"${DOMAIN}"/env.json -c)"; fi

# https://docs.aws.amazon.com/cli/latest/reference/lambda/create-function.html
aws lambda create-function \
  --function-name "plumbus_${DOMAIN}Handler" \
  --runtime "go1.x" \
  --role "${ROLE}" \
  --handler "main" \
  --description "${DESC}" \
  --zip-file "fileb://./main.zip" \
  --memory-size "${MEMORY}" \
  --timeout "${TIMEOUT}" \
  --environment "${ENV}";

printf "\n==> λƒ created!\n\n"