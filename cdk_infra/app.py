import os
from aws_cdk import (
    aws_lambda as lambda_,
    aws_iam as iam,
    aws_lambda_go as alg,
    aws_s3 as s3,
    aws_sqs as sqs,
    aws_s3_notifications as s3n,
    aws_lambda_event_sources as eventsources,
    core,
)

COMPONENT_MIDI_FILES_BUCKET = "component-midi-files"
MIDI_FILE_DROPOFF_BUCKET = "midi-file-dropoff"
CREATED_MP3_FILES_BUCKET = "created-mp3-files"

class MIDIStack(core.Stack):
    def __init__(self, app: core.App, id: str, **kwargs) -> None:
        super().__init__(app, id, **kwargs)


        # Policies
        s3_access_policy = iam.ManagedPolicy.from_managed_policy_arn(
            self, id="s3_access_policy", managed_policy_arn="arn:aws:iam::aws:policy/AmazonS3FullAccess")

        lambda_access_policy = iam.ManagedPolicy.from_managed_policy_arn(
            self, id="lambda_access_policy", managed_policy_arn="arn:aws:iam::aws:policy/AWSLambda_FullAccess")
        
        logs_policy = iam.ManagedPolicy.from_managed_policy_arn(
            self, id="logs_policy", managed_policy_arn="arn:aws:iam::aws:policy/CloudWatchLogsFullAccess")

        # Roles
        lambda_role = iam.Role(self, id="lambda_role", assumed_by=iam.ServicePrincipal(service="lambda.amazonaws.com"),
                               managed_policies=[s3_access_policy, lambda_access_policy, logs_policy],
            role_name=f"midi-to-mp3-lambda-role")

        # SQS
        conversion_sqs = sqs.Queue(self, id=f"conversion_sqs",
                                 queue_name=f"conversion_sqs",
                                 visibility_timeout=core.Duration.hours(12),
                                 retention_period=core.Duration.days(1))
        
        # S3
        
        midi_file_dropoff_bucket = s3.Bucket(self, id="midi_files_dropoff",
                                            bucket_name=MIDI_FILE_DROPOFF_BUCKET,
                                            auto_delete_objects=True,
                                            removal_policy=core.RemovalPolicy.DESTROY)
        
        created_mp3_files_bucket = s3.Bucket(self, id="created_mp3_files",
                                            bucket_name=CREATED_MP3_FILES_BUCKET,
                                            auto_delete_objects=True,
                                            removal_policy=core.RemovalPolicy.DESTROY)
        
        component_midi_files_bucket = s3.Bucket(self, id="component_midi_files",
                                            bucket_name=COMPONENT_MIDI_FILES_BUCKET,
                                            auto_delete_objects=True,
                                            removal_policy=core.RemovalPolicy.DESTROY)

        component_midi_files_bucket.add_event_notification(event=s3.EventType.OBJECT_CREATED,
                                                        dest=s3n.SqsDestination(conversion_sqs))

        # Lambdas
        lambda_code = lambda_.DockerImageCode.from_image_asset(directory='./midi_to_mp3_lambda/',
                                                               file="Dockerfile",
                                                               build_args={"AWS_ACCESS_KEY_ID": os.environ.get("AWS_ACCESS_KEY_ID"),
                                                                           "AWS_SECRET_ACCESS_KEY": os.environ.get("AWS_SECRET_ACCESS_KEY")
                                                                           })
        
        midi_to_mp3_lambda = lambda_.DockerImageFunction(self, id="midi_to_mp3_lambda",
                                         role=lambda_role,
                                         function_name="midi-to-mp3",
                                         memory_size=1024,
                                         timeout=core.Duration.minutes(5),
                                         code=lambda_code
                                         )

        midi_split_lambda = alg.GoFunction(self, id="midi_split_lambda",
                       entry="./midi_split_lambda/midi_split_lambda.go",
                       timeout=core.Duration.minutes(
                           15),
                       runtime=lambda_.Runtime.GO_1_X,
                       role=lambda_role,
                       function_name=f"midi-split-lambda",
                       memory_size=512,
                       bundling={
                           "environment": {
                               "GO111MODULE": "off"
                           }
                       })
        
        # Event Sources
        
        midi_to_mp3_lambda.add_event_source(
            eventsources.SqsEventSource(queue=conversion_sqs)
        )

        



app = core.App()
aws_account = core.Environment(account="463511281384", region="us-east-2")
MIDIStack(app, "MIDI-Part-Splitter", env=aws_account)
app.synth()

if __name__ == '__main__':
    print(os.path.abspath('./midi_to_mp3_lambda/'))