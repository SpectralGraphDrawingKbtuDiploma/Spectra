#!/usr/bin/env python3
import os
import sys
import glob
import argparse
import boto3
from botocore.exceptions import ClientError

def upload_to_s3(local_path, s3_directory):
    """
    Upload a file to the MinIO 'artifacts' bucket in the specified directory
    and save the S3 paths to result.txt

    :param local_path: Directory containing out.* files
    :param s3_directory: Directory in the S3 bucket
    :return: True if file was uploaded, else False
    """
    # Define the bucket name - this is now hardcoded as 'artifacts'
    bucket_name = 'artifacts'

    # Initialize the S3 client with MinIO endpoint
    s3_client = boto3.client(
        's3',
        endpoint_url='http://minio:9000',  # MinIO endpoint from docker-compose
        aws_access_key_id='minioadmin',    # Default MinIO access key
        aws_secret_access_key='minioadmin', # Default MinIO secret key
        config=boto3.session.Config(signature_version='s3v4'),
        verify=False  # Disable SSL verification for local development
    )

    # Find files matching pattern 'out.*'
    out_files = glob.glob(os.path.join(local_path, 'out.*'))

    if not out_files:
        print(f"Error: No files matching 'out.*' found in {local_path}")
        return False

    success = True
    # List to store successful uploads' paths
    uploaded_files = []

    for file_path in out_files:
        file_name = os.path.basename(file_path)
        s3_path = f"{s3_directory.rstrip('/')}/{file_name}"

        try:
            print(f"Uploading {file_path} to minio://{bucket_name}/{s3_path}")
            s3_client.upload_file(file_path, bucket_name, s3_path)
            full_s3_path = f"s3://{bucket_name}/{s3_path}"
            uploaded_files.append(full_s3_path)
            print(f"Successfully uploaded {file_name} to MinIO bucket '{bucket_name}'")
        except ClientError as e:
            print(f"Error uploading {file_name} to MinIO: {e}")
            success = False

    # Write the S3 paths to result.txt
    if uploaded_files:
        result_file_path = os.path.join(local_path, 'result.txt')
        try:
            with open(result_file_path, 'w') as result_file:
                for s3_file_path in uploaded_files:
                    result_file.write(f"{s3_file_path}\n")
            print(f"Successfully saved S3 paths to {result_file_path}")
        except Exception as e:
            print(f"Error saving result.txt: {e}")
            success = False

    return success

def main():
    parser = argparse.ArgumentParser(description='Upload out.* files to MinIO artifacts bucket')
    parser.add_argument('--local-path', required=True, help='Local directory containing out.* files')
    parser.add_argument('--s3-directory', required=True, help='Directory in the S3 bucket')

    args = parser.parse_args()

    if not os.path.isdir(args.local_path):
        print(f"Error: Local path {args.local_path} is not a directory or does not exist")
        sys.exit(1)

    success = upload_to_s3(args.local_path, args.s3_directory)

    if not success:
        sys.exit(1)

if __name__ == "__main__":
    main()
