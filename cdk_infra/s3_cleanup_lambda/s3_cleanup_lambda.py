from datetime import datetime, timedelta, timezone
import os

import boto3

NUM_WEEKS_TO_KEEP_FILES = int(os.environ.get("NUM_WEEKS_TO_KEEP_FILES", "1"))

COMPONENT_MIDI_FILES_BUCKET = "component-midi-files"
MIDI_FILE_DROPOFF_BUCKET = "midi-file-dropoff"
CREATED_MP3_FILES_BUCKET = "created-mp3-files"

buckets_to_clear = [COMPONENT_MIDI_FILES_BUCKET,
                    MIDI_FILE_DROPOFF_BUCKET,
                    CREATED_MP3_FILES_BUCKET]

def clear_out_s3_files(bucket):
    s3_client = boto3.client('s3')
    files = s3_client.list_objects_v2(
        Bucket=bucket
    )
    file_contents = files.get('Contents')
    now = datetime.now(tz=timezone.utc)
    n_weeks_ago = now + timedelta(weeks=-NUM_WEEKS_TO_KEEP_FILES)
    for f in file_contents:
        file_ts = f.get("LastModified")
        if n_weeks_ago > file_ts:
            k = f.get("Key")
            print(k, "will be deleted")
            s3_client.delete_object(Bucket=bucket,
                                    Key=k)

def handler(event=None, context=None):
    for bucket in buckets_to_clear:
        clear_out_s3_files(bucket)


if __name__ == '__main__':
    handler()
