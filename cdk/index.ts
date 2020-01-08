import { App } from '@aws-cdk/core';
import { LambdaStack } from './lambda-stack';

const app = new App();
new LambdaStack(app, 'CertBotStack');
app.synth();
