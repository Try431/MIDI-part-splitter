from shutil import make_archive
from aws_cdk import (
    aws_lambda as lambda_,
    aws_iam as iam,
    core,
)

DATAFRAME_TO_CSV_BUCKET = "kinetic-py2d2-df-csv-output"
REDSHIFT_TO_CSV_BUCKET = "redshift-timestmap-export-api-csv"


class MIDIStack(core.Stack):
    def __init__(self, app: core.App, id: str, **kwargs) -> None:
        super().__init__(app, id, **kwargs)


        # Policies
        s3_access_policy = iam.ManagedPolicy.from_managed_policy_arn(
            self, id="s3_access_policy", managed_policy_arn="arn:aws:iam::aws:policy/AmazonS3FullAccess")

        lambda_access_policy = iam.ManagedPolicy.from_managed_policy_arn(
            self, id="lambda_access_policy", managed_policy_arn="arn:aws:iam::aws:policy/AWSLambda_FullAccess")

        # Roles
        lambda_role = iam.Role(self, id="lambda_role", assumed_by=iam.ServicePrincipal(service="lambda.amazonaws.com"),
                               managed_policies=[s3_access_policy, lambda_access_policy],
            role_name=f"midi-to-mp3-lambda-role")

        # SQS
        # request_sqs = sqs.Queue(self, id=f"py2d2-sqs",
        #                         queue_name=f"py2d2-request-sqs",
        #                         visibility_timeout=core.Duration.hours(2),
        #                         retention_period=core.Duration.hours(1))

        # S3

        # the S3 bucket that holds the CSVs exported from Redshift
        # redshift_csv_bucket = s3.Bucket(self, id="redshift_csv_bucket",
        #                                 bucket_name=REDSHIFT_TO_CSV_BUCKET,
        #                                 auto_delete_objects=False)

        # MIDI to MP3 Lambda
        # subprocess.check_call([sys.executable, "-m", "pip", "install", "pandas==1.3.4", "--target", "raw_df_lambda/libs/"])

        # this is packaging up the lambda's external dependenices into a .zip file
        make_archive("midi_to_mp3_lambda/midi_to_mp3_lambda",
                     'zip', "./midi_to_mp3_lambda/")

        midi_to_mp3_lambda = lambda_.Function(self, id="midi_to_mp3_lambda",
                                         runtime=lambda_.Runtime.PYTHON_3_8,
                                         role=lambda_role,
                                         function_name="midi-to-mp3",
                                         memory_size=1024,
                                         timeout=core.Duration.minutes(15),
                                         handler="midi_to_mp3_lambda.handler",
                                         code=lambda_.Code.from_asset(
                                             "./midi_to_mp3_lambda/midi_to_mp3_lambda.zip")
                                         )


app = core.App()
aws_account = core.Environment(account="463511281384", region="us-east-2")
MIDIStack(app, "MIDI-Part-Splitter", env=aws_account)
app.synth()
