#!/bin/bash

cd infrastructure && cdk deploy --require-approval never "LinkShortenerFrontendStack"
FE_BUCKET_NAME=$(aws cloudformation describe-stacks --stack-name LinkShortenerFrontendStack --query "Stacks[0].Outputs[?OutputKey=='febucketname'].OutputValue" --output text)

cd ../frontend/dist/linkShortener && aws s3 rm --recursive s3://$FE_BUCKET_NAME && aws s3 sync . s3://$FE_BUCKET_NAME
cd ../../..