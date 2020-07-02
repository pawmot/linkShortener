#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { LinkShortenerFrontendStack } from '../lib/link-shortener-frontend-stack';
import {LinkShortenerDeployStack} from "../lib/link-shortener-deploy-stack";
import {LinkShortenerServiceStack} from "../lib/link-shortener-service-stack";

const app = new cdk.App();
new LinkShortenerFrontendStack(app, 'LinkShortenerFrontendStack');
new LinkShortenerDeployStack(app, "LinkShortenerDeployStack");
new LinkShortenerServiceStack(app, "LinkShortenerServiceStack");
