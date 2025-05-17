# S3 Backup Tool

A simple command-line tool to backup local directories to Amazon S3.

## Prerequisites

- Go 1.24 or later
- AWS credentials configured (via AWS CLI or environment variables)

## Installation

```bash
go mod tidy
go build
```

## Usage

```bash
./s3backup -source /path/to/source/directory -bucket your-bucket-name [-prefix optional/s3/prefix]
```

### Parameters

- `-source`: (Required) The local directory to backup
- `-bucket`: (Required) The S3 bucket name to upload to
- `-prefix`: (Optional) A prefix to add to all S3 keys

## AWS Credentials

The program uses the AWS SDK's default credential provider chain. You can configure your credentials in several ways:

1. Using AWS CLI: `aws configure`
2. Environment variables: `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`
3. Shared credentials file: `~/.aws/credentials`

## Example

```bash
./s3backup -source /home/user/documents -bucket my-backup-bucket -prefix backups/2024
```

This will upload all files from `/home/user/documents` to the S3 bucket `my-backup-bucket` under the prefix `backups/2024/`. 