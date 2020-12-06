package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"storj.io/uplink"
)

// BufferSize is the size of each file to read into memory at a time
const BufferSize = 1024

var (
	dryRun  = kingpin.Flag("dry-run", "dry run").Envar("DRY_RUN").Bool()
	access  = kingpin.Flag("plugin-access", "tardigrade access").Envar("PLUGIN_ACCESS").Required().String()
	bucket  = kingpin.Flag("plugin-bucket", "tardigrade bucket").Envar("PLUGIN_BUCKET").Required().String()
	source  = kingpin.Flag("plugin-source", "source pattern").Envar("PLUGIN_SOURCE").Required().String()
	exclude = kingpin.Flag("plugin-exclude", "exclude pattern").Envar("PLUGIN_EXCLUDE").Default().Strings()
	target  = kingpin.Flag("plugin-target", "target directory").Envar("PLUGIN_TARGET").Required().String()
)

func main() {
	kingpin.Parse()
	userAccess, err := uplink.ParseAccess(*access)
	if err != nil {
		logrus.WithError(err).Fatalln("unable to parse access")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	project, err := uplink.OpenProject(ctx, userAccess)
	if err != nil {
		logrus.WithError(err).Fatalln("unable to open project")
	}
	defer project.Close()

	matches, err := matches(*source, *exclude)
	if err != nil {
		logrus.WithError(err).Fatalln("unable to build matches file list")
	}
	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		if stat.IsDir() {
			continue
		}

		key := filepath.Join(*target, match)

		logrus.WithFields(logrus.Fields{
			"name":   match,
			"bucket": *bucket,
			"target": key,
		}).Info("uploading file")

		f, err := os.Open(match)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"file": match,
			}).Fatalln("problem opening file")
		}
		defer f.Close()

		if !(*dryRun) {
			buffer := make([]byte, BufferSize)

			upload, err := project.UploadObject(ctx, *bucket, key, nil)
			if err != nil {
				logrus.WithError(err).Fatalln("unable to create upload")
			}

			for {
				_, err := f.Read(buffer)
				if err != nil {
					if err != io.EOF {
						logrus.WithError(err).Fatalln("unable to read file")
					}
					_, err = upload.Write(buffer)
					if err != nil {
						upload.Abort()
						logrus.WithError(err).Fatalln("unable to upload file")
					}
					err = upload.Commit()
					if err != nil {
						logrus.WithError(err).Fatalln("unable to commit upload")
					}
					break
				}
			}
		} else {
			logrus.Info("skipping file upload... dry run enabled")
		}
	}
}

// matches is a helper function that returns a list of all files matching the
// included Glob pattern, while excluding all files that matche the exclusion
// Glob pattners.
func matches(include string, exclude []string) ([]string, error) {
	matches, err := zglob.Glob(include)
	if err != nil {
		return nil, errors.Wrap(err, "glob failed")
	}
	if len(exclude) == 0 {
		return matches, nil
	}

	// find all files that are excluded and load into a map. we can verify
	// each file in the list is not a member of the exclusion list.
	excludem := map[string]bool{}
	for _, pattern := range exclude {
		excludes, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range excludes {
			excludem[match] = true
		}
	}

	var included []string
	for _, include := range matches {
		_, ok := excludem[include]
		if ok {
			continue
		}
		included = append(included, include)
	}
	return included, nil
}
