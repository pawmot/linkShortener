import * as cdk from '@aws-cdk/core';
import * as s3 from '@aws-cdk/aws-s3';
import {CfnOutput} from "@aws-cdk/core";

export class LinkShortenerFrontendStack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        let feBucket = new s3.Bucket(this, "linkShortener-frontend", {
            publicReadAccess: true,
            websiteIndexDocument: "index.html"
        });

        new CfnOutput(this, "fe-bucket-name", {
            value: feBucket.bucketName
        });

        new CfnOutput(this, "fe-bucket-base-url", {
            value: feBucket.bucketWebsiteUrl
        });
    }
}
