package main

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jawher/mow.cli"
	"github.com/klauspost/compress/snappy"
	"github.com/rlmcpherson/s3gof3r"
	"net/http"
)

func main() {
	app := cli.App("neo4j-hot-backup", "Backup neo4j to S3")

	s3domain := app.String(cli.StringOpt{
		Name:   "s3domain",
		Desc:   "s3 domain",
		EnvVar: "S3_DOMAIN",
		Value:  "s3-eu-west-1.amazonaws.com",
	})

	s3bucket := app.String(cli.StringOpt{
		Name:   "bucket",
		Desc:   "s3 bucket name",
		EnvVar: "S3_BUCKET",
		Value:  "com.ft.coco-neo4j-backup",
	})

	s3dir := app.String(cli.StringOpt{
		Name:   "base-dir",
		Desc:   "s3 base directory name",
		EnvVar: "S3_DIR",
		Value:  "/backups/",
	})

	accessKey := app.String(cli.StringOpt{
		Name:   "aws_access_key_id",
		Desc:   "AWS Access key id",
		EnvVar: "AWS_ACCESS_KEY_ID",
	})

	secretKey := app.String(cli.StringOpt{
		Name:      "aws_secret_access_key",
		Desc:      "AWS secret access key",
		EnvVar:    "AWS_SECRET_ACCESS_KEY",
		HideValue: true,
	})

	dir := app.String(cli.StringOpt{
		Name:   "dir",
		Desc:   "backup location",
		EnvVar: "BACKUP_DIR",
		Value:  "/backup",
	})

	app.Command("backup", "backup a neo backup to S3", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			n := newNeoBackup(*s3bucket, *s3dir, *s3domain, *accessKey, *secretKey)
			if err := n.backupDirectory(*dir); err != nil {
				log.Fatalf("backup failed : %v\n", err)
			}
		}
	})

	app.Command("restore", "restore a neo backup from S3", func(cmd *cli.Cmd) {
		dateDir := cmd.String(cli.StringArg{
			Name: "DATE",
			Desc: "Date to restore backup from",
		})
		cmd.Action = func() {
			n := newNeoBackup(*s3bucket, *s3dir, *s3domain, *accessKey, *secretKey)
			if err := restoreDirectory(n, *dateDir, *dir); err != nil {
				log.Fatalf("restore failed : %v\n", err)
			}
		}
	})

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

type neoBackup struct {
	s3bucket string
	s3dir    string
	s3       *s3gof3r.S3
}

type nb interface {
	readS3(dateDir string) (string, io.ReadCloser, http.Header, error)
}

func newNeoBackup(s3bucket, s3dir, s3domain, accessKey, secretKey string) *neoBackup {
	return &neoBackup{
		s3bucket,
		s3dir,
		s3gof3r.New(
			s3domain,
			s3gof3r.Keys{
				AccessKey: accessKey,
				SecretKey: secretKey,
			},
		),
	}
}

func (n *neoBackup) readS3(dataDir string) (path string, http io.ReadCloser, headers http.Header, err error) {
	path = filepath.Join(n.s3dir, dataDir+".tar.snappy")
	http, headers, err = n.s3.Bucket(n.s3bucket).GetReader(path, nil)
	return path, http, headers, err
}

func restoreDirectory(n nb, dateDir, dir string) error {

	path, br, _, err := n.readS3(dateDir)
	if err != nil {
		return err
	}
	defer br.Close()
	sr := snappy.NewReader(br)
	tr := tar.NewReader(sr)

	log.Printf("[INFO] Restoring '%s' to '%s'\n", path, dir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fullPath := filepath.Join(dir, header.Name)
			err = os.MkdirAll(filepath.Dir(fullPath), 0777)
			if err != nil {
				log.Fatal(err)
			}
			file, e := os.Create(fullPath)

			if e != nil {
				log.Fatal(e)
			}
			defer file.Close()
			io.Copy(file, tr)
		}
	}
	log.Printf("[INFO] Restore of '%s' to '%s' complete\n", path, dir)
	return nil
}

func (n *neoBackup) backupDirectory(dir string) error {

	dateDir := formattedNow()
	b := n.s3.Bucket(n.s3bucket)
	path := filepath.Join(n.s3dir, dateDir+".tar.snappy")
	bw, err := b.PutWriter(path, nil, nil)
	if err != nil {
		return err
	}
	defer bw.Close()

	sw := snappy.NewBufferedWriter(bw)
	tw := tar.NewWriter(sw)
	defer sw.Close()
	defer tw.Close()

	log.Printf("[INFO] Backing up directory %s to %s\n", dir, path)
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		header, err := tar.FileInfoHeader(info, path[len(dir):])
		header.Name = path[len(dir):]

		if err != nil {
			return err
		}
		if err = tw.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		defer file.Close()
		_, err = io.Copy(tw, file)
		if err != nil {
			log.Fatal(err)
		}
		return err
	})
	if err != nil {
		return err
	}
	log.Printf("[INFO] Backup of directory %s to %s is complete \n", dir, path)
	return nil
}

func formattedNow() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05")
}
