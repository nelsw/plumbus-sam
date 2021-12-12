#!/usr/bin/env bash
#
# This script validates an AWS Lambda Function code and configuration by providing valid variable parameters.
printf "\n==> updating an AWS λƒ...\n"

# The following conditions validation command variable by asserting and defaulting values.
if [ -z "${DOMAIN}" ]; then echo "ERROR: Set DOMAIN to the entity identity of the λƒ to build"; exit 1; fi
if [ -z "${ROLE}" ]; then echo "ERROR: Set ROLE to the AWS IAM Role of the λƒ to build"; exit 1; fi
if [ -z "${TIMEOUT}" ]; then TIMEOUT="30"; fi
if [ -z "${MEMORY}" ]; then MEMORY="512"; fi
if [ -z "${DESC}" ]; then DESC="null"; fi
if [ -f test/"${DOMAIN}"/env.json ]; then ENV="$(jq '.' test/"${DOMAIN}"/env.json -c)"; fi

# https://docs.aws.amazon.com/cli/latest/reference/lambda/update-function-configuration.html
printf "\n==> updating the configuration...\n"
aws lambda update-function-configuration \
  --function-name "${DOMAIN}Handler" \
  --role "${ROLE}" \
  --description "${DESC}" \
  --timeout "${TIMEOUT}" \
  --memory-size "${MEMORY}" \
  --environment "${ENV}";
printf "==> configuration updated!\n"

# https://docs.aws.amazon.com/cli/latest/reference/lambda/update-function-code.html
printf "\n==> updating the code...\n"
aws lambda update-function-code \
  --function-name "${DOMAIN}Handler" \
  --zip-file fileb://./main.zip
printf "==> code updated!\n"

printf "\n==> λƒ updated!\n\n"
