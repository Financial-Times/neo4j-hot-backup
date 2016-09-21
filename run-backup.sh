#!/usr/bin/env bash

TAR_UP=neo4j-$ENV-$(date +"%Y%m%d%H%M").tar.gz

tar -czf /$TAR_UP /upload

aws s3 cp /$TAR_UP s3://$AWS_BUCKET_NAME/$ENV/