import events = require('@aws-cdk/aws-events');
import targets = require('@aws-cdk/aws-events-targets');
import lambda = require('@aws-cdk/aws-lambda');
import iam = require('@aws-cdk/aws-iam');
import s3 = require('@aws-cdk/aws-s3');
import cdk = require('@aws-cdk/core');

export class LambdaStack extends cdk.Stack {
    constructor(app: cdk.App, id: string) {
        super(app, id);

        const OutputBucketName = app.node.tryGetContext('OutputBucketName')
        if (!OutputBucketName) {
            throw new Error('Context variable OutputBucketName is required!')
        }

        const Domains = app.node.tryGetContext('Domains')
        if (!Domains) {
            throw new Error('Context variable Domains is requried!')
        }

        const AcmeServer = app.node.tryGetContext('AcmeServer')
        if (!AcmeServer) {
            throw new Error('Context variable AcmeServer is required!')
        }

        const AccountEmail = app.node.tryGetContext('AccountEmail')
        if (!AccountEmail) {
            throw new Error('Context variable AccountEmail is required!')
        }


        const outputBucket = new s3.Bucket(this, 'CertBotOutputBucket', {
            bucketName: OutputBucketName
        });

        const executionRole = new iam.Role(this, 'CertBotLambdaExecutionRole', {
            assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com')
        });
        executionRole.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            resources: ['*'],
            actions: [
                's3:PutObject',
                'secretsmanager:CreateSecret',
                'route53:ListHostedZonesByName'
            ]
        }))
        executionRole.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            resources: [`arn:aws:secretsmanager:${this.region}:${this.account}:secret:acme/*`],
            actions: [
                'secretsmanager:GetSecretValue',
                'secretsmanager:PutSecretValue'
            ]
        }))
        executionRole.addToPolicy(new iam.PolicyStatement({
            effect: iam.Effect.ALLOW,
            resources: [
                'arn:aws:route53:::hostedzone/*',
                'arn:aws:route53:::change/*'
            ],
            actions: [
                'route53:GetChange',
                'route53:ChangeResourceRecordSets',
                'route53:ListResourceRecordSets'
            ]
        }))


        for (let domain of Domains) {

            const certRenewalLambda = new lambda.Function(this, `CertBotLambda-${domain}`, {
                code: lambda.Code.fromAsset('./cert-bot.zip'),
                handler: 'cert-bot',
                timeout: cdk.Duration.seconds(300),
                runtime: lambda.Runtime.GO_1_X,
                environment: {
                    ACME_SERVER: AcmeServer,
                    ACCOUNT_EMAIL: AccountEmail,
                    OUTPUT_BUCKET: outputBucket.bucketName,
                    FQDN: domain
                },
                role: executionRole
            });
    
            const rule = new events.Rule(this, `CertBotLambdaScheduler-${domain}`, {
                schedule: events.Schedule.expression("rate(7 days)")
            });

            rule.addTarget(new targets.LambdaFunction(certRenewalLambda));
        }
    }
}
