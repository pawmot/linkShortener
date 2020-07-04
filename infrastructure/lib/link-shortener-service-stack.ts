import * as cdk from "@aws-cdk/core";
import * as lambda from "@aws-cdk/aws-lambda";
import * as apigw from "@aws-cdk/aws-apigatewayv2";
import * as ddb from "@aws-cdk/aws-dynamodb";
import * as s3 from "@aws-cdk/aws-s3";
import * as iam from "@aws-cdk/aws-iam";
import * as ec2 from "@aws-cdk/aws-ec2";
import * as elasticache from "@aws-cdk/aws-elasticache";

export class LinkShortenerServiceStack extends cdk.Stack {
    constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
        super(scope, id, props);

        const params = handleParameters(this);

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

        let counterTable = new ddb.Table(this, "linkShortener_counter", {
            partitionKey: {
                name: "cnt_name",
                type: ddb.AttributeType.STRING
            },
            readCapacity: 1,
            writeCapacity: 1,
            tableName: "linkShortener_counter"
        });

        let vpc = new ec2.Vpc(this,
            "linkShortenerVpc",
            {
                cidr: "10.0.0.0/16"
            });

        let sg = new ec2.SecurityGroup(this, "cacheSecGroup", {
            vpc: vpc,
            securityGroupName: "cacheSecGroup"
        });

        sg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(11211), "memcached");

        let cacheSubnetGroup = new elasticache.CfnSubnetGroup(
            this,
            "linksCacheSubnetGroup",
            {
                cacheSubnetGroupName: "cachePrivateSG",
                subnetIds: vpc.privateSubnets.map(s => s.subnetId),
                description: "Link Shortener Cache Subnet Group"
            }
        );

        let cache = new elasticache.CfnCacheCluster(this, "linksCache", {
            engine: "memcached",
            cacheNodeType: "cache.t3.micro",
            numCacheNodes: 3,
            cacheSubnetGroupName: cacheSubnetGroup.cacheSubnetGroupName,
            vpcSecurityGroupIds: [sg.securityGroupId]
        });
        cache.addDependsOn(cacheSubnetGroup);

        let deployBucket = s3.Bucket.fromBucketName(this, "deployBucket", params.deployBucketName);

        let registerLinkFn = new lambda.Function(this, "registerLink", {
            runtime: lambda.Runtime.GO_1_X,
            code: lambda.Code.fromBucket(deployBucket, params.registerLinkFilename),
            handler: "main",
            currentVersionOptions: {
                codeSha256: params.registerLinkSha256
            },
            environment: {
                "LINKS_TABLE_NAME": linksTable.tableName,
                "URL_IDX_NAME": indexName,
                "COUNTER_TABLE_NAME": counterTable.tableName,
                "MEMCACHED_ADDR": "lin-li-qgvq1ji9fty8.fbscbm.0001.euw1.cache.amazonaws.com:11211"
            },
            vpc: vpc,
            vpcSubnets: {
                subnets: vpc.privateSubnets
            }
        });
        registerLinkFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [linksTable.tableArn],
            effect: iam.Effect.ALLOW,
            actions: ["dynamodb:PutItem"]
        }));
        registerLinkFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [`${linksTable.tableArn}/index/${indexName}`],
            effect: iam.Effect.ALLOW,
            actions: ["dynamodb:Query"]
        }));
        registerLinkFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [counterTable.tableArn],
            effect: iam.Effect.ALLOW,
            actions: ["dynamodb:UpdateItem"]
        }));

        let registerLinkIntegration = new apigw.LambdaProxyIntegration({
            handler: registerLinkFn
        });

        let redirectFn = new lambda.Function(this, "redirect", {
            runtime: lambda.Runtime.GO_1_X,
            code: lambda.Code.fromBucket(deployBucket, params.redirectFilename),
            handler: "main",
            currentVersionOptions: {
                codeSha256: params.redirectSha256
            },
            environment: {
                "TABLE_NAME": linksTable.tableName,
                "MEMCACHED_ADDR": "lin-li-qgvq1ji9fty8.fbscbm.0001.euw1.cache.amazonaws.com:11211"
            },
            vpc: vpc,
            vpcSubnets: {
                subnets: vpc.privateSubnets
            }
        });
        redirectFn.addToRolePolicy(new iam.PolicyStatement({
            resources: [linksTable.tableArn],
            effect: iam.Effect.ALLOW,
            actions: ["dynamodb:GetItem"]
        }));

        let redirectIntegration = new apigw.LambdaProxyIntegration({
            handler: redirectFn
        });

        let api = new apigw.HttpApi(this, "LinkShortenerApi", {
            corsPreflight: {
                allowOrigins: [params.frontendOrigin],
                allowMethods: [
                    apigw.HttpMethod.POST,
                    apigw.HttpMethod.GET,
                    apigw.HttpMethod.OPTIONS],
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

        new cdk.CfnOutput(this, "apiUrl", {
            value: api.url ?? "NO URL!!"
        })
    }
}

interface Parameters {
    deployBucketName: string;
    registerLinkFilename: string;
    registerLinkSha256: string;
    redirectFilename: string;
    redirectSha256: string;
    frontendOrigin: string;
}

function handleParameters(c: cdk.Construct): Parameters {

    let deployBucketNameParam = new cdk.CfnParameter(c, "deployBucketName", {
        type: "String",
        description: "Name of the bucket used to deploy Lambda code"
    });

    let registerLinkFilenameParam = new cdk.CfnParameter(c, "registerLinkFilename", {
        type: "String",
        description: "Name of the registerLink function code file"
    });

    let registerLinkSha256Param = new cdk.CfnParameter(c, "registerLinkSha256", {
        type: "String",
        description: "SHA256 of the registerLink zip file"
    });

    let redirectFilenameParam = new cdk.CfnParameter(c, "redirectFilename", {
        type: "String",
        description: "Name of the redirect function code file"
    });

    let redirectSha256Param = new cdk.CfnParameter(c, "redirectSha256", {
        type: "String",
        description: "SHA256 of the redirect zip file"
    });

    let frontendOriginParam = new cdk.CfnParameter(c, "frontendOrigin", {
        type: "String",
        description: "Origin of the frontend app (for CORS)"
    });

    return {
        deployBucketName: deployBucketNameParam.valueAsString,
        registerLinkFilename: registerLinkFilenameParam.valueAsString,
        registerLinkSha256: registerLinkSha256Param.valueAsString,
        redirectFilename: redirectFilenameParam.valueAsString,
        redirectSha256: redirectSha256Param.valueAsString,
        frontendOrigin: frontendOriginParam.valueAsString
    }
}
