import * as cdk from "@aws-cdk/core";
import {CfnOutput} from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as apigw from "@aws-cdk/aws-apigatewayv2";
import {HttpMethod} from "@aws-cdk/aws-apigatewayv2";
import * as ddb from "@aws-cdk/aws-dynamodb";
import * as s3 from "@aws-cdk/aws-s3";
import * as iam from "@aws-cdk/aws-iam";
import {Effect} from "@aws-cdk/aws-iam";

export class LinkShortenerServiceStack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        let deployBucketNameParam = new cdk.CfnParameter(this, "deployBucketName", {
            type: "String",
            description: "Name of the bucket used to deploy Lambda code"
        });

        let registerLinkFilenameParam = new cdk.CfnParameter(this, "registerLinkFilename", {
            type: "String",
            description: "Name of the registerLink function code file"
        });

        let registerLinkSha256Param = new cdk.CfnParameter(this, "registerLinkSha256", {
            type: "String",
            description: "SHA256 of the registerLink zip file"
        });

        let redirectFilenameParam = new cdk.CfnParameter(this, "redirectFilename", {
            type: "String",
            description: "Name of the redirect function code file"
        });

        let redirectSha256Param = new cdk.CfnParameter(this, "redirectSha256", {
            type: "String",
            description: "SHA256 of the redirect zip file"
        });

        let frontendOriginParam = new cdk.CfnParameter(this, "frontendOrigin", {
            type: "String",
            description: "Origin of the frontend app (for CORS)"
        });

        let deployBucketName = deployBucketNameParam.valueAsString;
        let registerLinkFilename = registerLinkFilenameParam.valueAsString;
        let registerLinkSha256 = registerLinkSha256Param.valueAsString;
        let redirectFilename = redirectFilenameParam.valueAsString;
        let redirectSha256 = redirectSha256Param.valueAsString;
        let frontendOrigin = frontendOriginParam.valueAsString;

        let linksTable = new ddb.Table(this, "linkShortener_links", {
            partitionKey: {
                name: "linkKey",
                type: ddb.AttributeType.STRING
            },
            readCapacity: 1,
            writeCapacity: 1,
            tableName: "linkShortener_links",
        });

        let indexName = "linkShortener_links_by_url";
        linksTable.addGlobalSecondaryIndex({
            indexName: indexName,
            partitionKey: {
                name: "url",
                type: ddb.AttributeType.STRING
            },
            readCapacity: 1,
            writeCapacity: 1
        });

        let deployBucket = s3.Bucket.fromBucketName(this, "deployBucket", deployBucketName);

        let registerLinkFn = new lambda.Function(this, "registerLink", {
            runtime: lambda.Runtime.GO_1_X,
            code: lambda.Code.fromBucket(deployBucket, registerLinkFilename),
            handler: "main",
            currentVersionOptions: {
                codeSha256: registerLinkSha256
            },
            environment: {
                "TABLE_NAME": linksTable.tableName,
                "URL_IDX_NAME": indexName
            }
        });
        registerLinkFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [linksTable.tableArn],
            effect: Effect.ALLOW,
            actions: ["dynamodb:PutItem"]
        }));

        let registerLinkIntegration = new apigw.LambdaProxyIntegration({
            handler: registerLinkFn
        });

        let redirectFn = new lambda.Function(this, "redirect", {
            runtime: lambda.Runtime.GO_1_X,
            code: lambda.Code.fromBucket(deployBucket, redirectFilename),
            handler: "main",
            currentVersionOptions: {
                codeSha256: redirectSha256
            },
            environment: {
                "TABLE_NAME": linksTable.tableName
            }
        });
        redirectFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [linksTable.tableArn],
            effect: Effect.ALLOW,
            actions: ["dynamodb:GetItem"]
        }));

        let redirectIntegration = new apigw.LambdaProxyIntegration({
            handler: redirectFn
        });

        let api = new apigw.HttpApi(this, "LinkShortenerApi", {
            corsPreflight: {
                allowOrigins: [frontendOrigin],
                allowMethods: [HttpMethod.POST, HttpMethod.GET, HttpMethod.OPTIONS],
                allowHeaders: ["*"],
                allowCredentials: false
            }
        });

        api.addRoutes({
            path: "/registerLink",
            methods: [apigw.HttpMethod.POST],
            integration: registerLinkIntegration
        });

        api.addRoutes({
            path: "/{key}",
            methods: [apigw.HttpMethod.GET],
            integration: redirectIntegration,
        });

        new CfnOutput(this, "apiUrl", {
            value: api.url ?? "NO URL!!"
        })
    }
}