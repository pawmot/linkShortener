#!/bin/bash

cd infrastructure && cdk deploy --require-approval never "LinkShortenerDeployStack"
DEPLOY_BUCKET_NAME=$(aws cloudformation describe-stacks --stack-name LinkShortenerDeployStack --query "Stacks[0].Outputs[?OutputKey=='deploybucketname'].OutputValue" --output text)
FE_BUCKET_WEBSITE_URL=$(aws cloudformation describe-stacks --stack-name LinkShortenerFrontendStack --query "Stacks[0].Outputs[?OutputKey=='febucketbaseurl'].OutputValue" --output text)

cd ../service/registerLink
GOOS=linux GOARCH=amd64 go build main.go
RL_HASH=$(sha1sum main | awk '{ print $1 }')
RL_FILENAME=deploy_rl_$RL_HASH.zip
zip $RL_FILENAME main
aws s3 cp $RL_FILENAME s3://$DEPLOY_BUCKET_NAME
RL_ZIP_SHA256=$(sha256sum $RL_FILENAME | awk '{ print $1 }')
rm $RL_FILENAME
rm main
cd ../

cd redirect
GOOS=linux GOARCH=amd64 go build main.go
RED_HASH=$(sha1sum main | awk '{ print $1 }')
RED_FILENAME=deploy_red_$RED_HASH.zip
zip $RED_FILENAME main
aws s3 cp $RED_FILENAME s3://$DEPLOY_BUCKET_NAME
RED_ZIP_SHA256=$(sha256sum $RED_FILENAME | awk '{ print $1 }')
rm $RED_FILENAME
rm main

cd ../../infrastructure && cdk deploy \
--require-approval never "LinkShortenerServiceStack" \
--parameters deployBucketName=$DEPLOY_BUCKET_NAME \
--parameters registerLinkFilename=$RL_FILENAME \
--parameters registerLinkSha256=$RL_ZIP_SHA256 \
--parameters redirectFilename=$RED_FILENAME \
--parameters redirectSha256=$RED_ZIP_SHA256 \
--parameters frontendOrigin=$FE_BUCKET_WEBSITE_URL
