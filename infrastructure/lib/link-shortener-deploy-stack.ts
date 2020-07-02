import * as cdk from '@aws-cdk/core';
import * as s3 from '@aws-cdk/aws-s3';
import {CfnOutput, Duration} from "@aws-cdk/core";

export class LinkShortenerDeployStack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        let deployBucket = new s3.Bucket(this, "linkShortener-deploy", {
            lifecycleRules: [
                {
                    expiration: Duration.days(1)
                }
            ]
        });

        new CfnOutput(this, "deploy-bucket-name", {
            value: deployBucket.bucketName
        });
    }
}