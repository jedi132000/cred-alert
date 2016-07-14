package queue

import (
	"archive/zip"
	"cred-alert/github"
	"cred-alert/metrics"
	"cred-alert/notifications"
	"cred-alert/scanners"
	"cred-alert/scanners/file"
	"cred-alert/sniff"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pivotal-golang/lager"
)

type RefScanJob struct {
	RefScanPlan
	client            github.Client
	sniff             sniff.SniffFunc
	notifier          notifications.Notifier
	emitter           metrics.Emitter
	credentialCounter metrics.Counter
}

func NewRefScanJob(plan RefScanPlan, client github.Client, sniff sniff.SniffFunc, notifier notifications.Notifier, emitter metrics.Emitter) *RefScanJob {
	credentialCounter := emitter.Counter("cred_alert.violations")

	job := &RefScanJob{
		RefScanPlan:       plan,
		client:            client,
		sniff:             sniff,
		notifier:          notifier,
		emitter:           emitter,
		credentialCounter: credentialCounter,
	}

	return job
}

func (j *RefScanJob) Run(logger lager.Logger) error {
	logger = logger.Session("ref-scan", lager.Data{
		"owner":      j.Owner,
		"repository": j.Repository,
		"ref":        j.Ref,
	})

	downloadURL, err := j.client.ArchiveLink(logger, j.Owner, j.Repository)
	if err != nil {
		logger.Error("Error getting download url", err)
	}

	archiveFile, err := downloadArchive(logger, downloadURL)
	if err != nil {
		logger.Error("Error downloading archive", err)
		return err
	}
	defer os.Remove(archiveFile.Name())
	defer archiveFile.Close()

	archiveReader, err := zip.OpenReader(archiveFile.Name())
	if err != nil {
		logger.Error("Error unzipping archive", err)
		return err
	}
	defer archiveReader.Close()

	for _, f := range archiveReader.File {
		unzippedReader, err := f.Open()
		if err != nil {
			logger.Error("Error reading archive", err)
			continue
		}
		defer unzippedReader.Close()

		bufioScanner := file.NewReaderScanner(unzippedReader, f.Name)
		handleViolation := j.createHandleViolation(logger, j.Ref, j.Owner+"/"+j.Repository)

		err = j.sniff(logger, bufioScanner, handleViolation)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadArchive(logger lager.Logger, link *url.URL) (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "downloaded-git-archive")
	if err != nil {
		logger.Error("Error creating archive temp file", err)
		return nil, err
	}

	resp, err := http.Get(link.String())
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return nil, err
	}

	return tempFile, nil
}

func (j *RefScanJob) createHandleViolation(logger lager.Logger, ref string, repoName string) func(scanners.Line) error {
	return func(line scanners.Line) error {
		logger.Info("found-credential", lager.Data{
			"path":        line.Path,
			"line-number": line.LineNumber,
			"ref":         ref,
		})

		err := j.notifier.SendNotification(logger, repoName, ref, line)
		if err != nil {
			return err
		}

		j.credentialCounter.Inc(logger)

		return nil
	}
}