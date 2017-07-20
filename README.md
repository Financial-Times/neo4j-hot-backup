## Neo4j Hot Backup
This app backs up any directory to S3, including adding the files to a tarball and compressing it using Snappy compression.

### Building
Remote builds are produced automatically on DockerHub by any commit to the master branch or any tag.

If you wish to build locally:
- Go app: `go build`
- Docker build: `docker build -t coco/neo4j-hot-backup .`

### Performing a backup

- &lt;BACKUP_DIRECTORY&gt;: location to backup to - the standard location is `/vol/neo4j/data/databases/graph.db`

```
docker run --rm \
--env AWS_ACCESS_KEY_ID=$(/usr/bin/etcdctl get /ft/_credentials/aws/aws_access_key_id) \
--env AWS_SECRET_ACCESS_KEY=$(/usr/bin/etcdctl get /ft/_credentials/aws/aws_secret_access_key) \
--env S3_BUCKET=com.ft.coco-neo4j-backup \
--env S3_DIR=$(/usr/bin/etcdctl get /ft/config/environment_tag) \
-v <BACKUP_DIRECTORY>:/backup \
coco/neo4j-hot-backup
```

### Performing a restore
- Make sure that the database has been shut down before moving/restoring the data.

```
fleetctl stop neo4j@{1..3}.service
```

- Before starting the restore, either move or delete the existing `/vol/neo4j/data/databases/graph.db` directory.
- If you want to move and keep a copy of the current data, check there is sufficient disk space available.

```
df -h
sudo mv /vol/neo4j/data/databases/graph.db /vol/neo4j/data/databases/graph.db.`date +%F`

OR

sudo rm -rf /vol/neo4j/data/databases/graph.db
```

- Start the restore using `neo4j-hot-backup`.
- &lt;ENVIRONMENT TAG&gt;: The environment that you'll be restoring the backup from.
- &lt;RESTORE_DIRECTORY&gt;: location to restore to - the standard location is `/vol/neo4j/data/databases/graph.db`
- Date (2016-09-23T14-30-11): The timestamp of the backup to restore.

```
docker run --rm \g
--env AWS_ACCESS_KEY_ID=$(/usr/bin/etcdctl get /ft/_credentials/aws/aws_access_key_id) \
--env AWS_SECRET_ACCESS_KEY=$(/usr/bin/etcdctl get /ft/_credentials/aws/aws_secret_access_key) \
--env S3_BUCKET=com.ft.coco-neo4j-backup \
--env S3_DIR=<ENVIRONMENT_TAG> \
-v <RESTORE_DIRECTORY>:/backup \
coco/neo4j-hot-backup ./neo4j-hot-backup restore 2016-09-23T14-30-11
```

- Start neo4j back up:

```
fleetctl start neo4j@{1..3}.service
```

- Once started, the indexes need to be rebuilt.  This can take a while and will look like it's stuck at `Initialising metrics...`


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
