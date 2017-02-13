## Neo4j Hot Backup
This app backs up any directory to S3, including adding the files to a tarball and compressing it using Snappy compression.

### Building
Remote builds are produced automatically on DockerHub by any commit to the master branch or any tag.

If you wish to build locally:
- Go app: `go build`
- Docker build: `docker build -t coco/neo4j-hot-backup .`

### Running

To backup

    docker run --rm \
    --env AWS_ACCESS_KEY_ID=<AWS ACCESS KEY> \
    --env AWS_SECRET_ACCESS_KEY=<AWS SECRET KEY> \
    --env S3_BUCKET=com.ft.coco-neo4j-backup \
    --env S3_DIR=<ENVIRONMENT TAG> \
    -v <backup folder>:/backup \
    coco/neo4j-hot-backup

To restore

    docker run --rm \
    --env AWS_ACCESS_KEY_ID=<AWS ACCESS KEY> \
    --env AWS_SECRET_ACCESS_KEY=<AWS SECRET KEY> \
    --env S3_BUCKET=com.ft.coco-neo4j-backup \
    --env S3_DIR=<ENVIRONMENT TAG> \
    -v <backup folder>:/backup \
    coco/neo4j-hot-backup ./neo4j-hot-backup restore 2016-09-23T14-30-11

- &lt;ENVIRONMENT TAG&gt; = The environment that you'll be restoring the backup from.
- &lt;backup folder&gt; = The local folder that you'll be restoring to.
- Date (2016-09-23T14-30-11): The timestamp of the backup to restore.

### Restoring in the cluster
- Make sure that the database has been shut down before moving/restoring the data.
- When restarting, the indexes need to be rebuilt.  This can take a while and will look like it's stuck at `Initialising metrics...`

### Testing for developers
When making changes, follow the testing procedure below to ensure that the base functionality still work.
- Run a Neo4j hot backup to produce a backup folder
- Run the backup command of this app to upload the folder to S3
 - Expected output:
 - `2016/09/23 14:30:00 [INFO] Backing up directory /backup to backup/2016-09-23T14-30-11.tar.snappy`
- Run the restore command of this app to download the backup from S3 to a new location.
 - Expected output: 
 - `2016/09/23 14:50:00 [INFO] Restoring backup/2016-09-23T14-30-11.tar.snappy to /backup`
 - `2016/09/23 14:50:00 [INFO] Restore of backup/2016-09-23T14-30-11.tar.snappy to /backup complete`
