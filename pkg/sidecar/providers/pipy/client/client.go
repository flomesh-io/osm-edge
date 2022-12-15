// Package client implements the PipyRepo struct.
package client

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// relativeAPIath is default Pipy api relative path
	relativeAPIath = "api"
	// relativeRepoPath is default Pipy repo api relative path
	relativeRepoPath = "repo"
	// relativeRepoPath is default Pipy repo files api relative path
	relativeRepoFilePath = "repo-files"

	apiVersion1 = "v1"

	defaultHTTPSchema = "http"
)

type repoAPIURI struct {
	serverAddr   string
	serverPort   uint16
	schema       string
	version      string
	apiURI       string
	repoURI      string
	repoFilesURI string

	baseRepoURI      string
	baseRepoFilesURI string
}

// newRepoAPIURI creates a Repo Api URIs
func newRepoAPIURI(serverAddr string, serverPort uint16) *repoAPIURI {
	return (&repoAPIURI{
		serverAddr:   serverAddr,
		serverPort:   serverPort,
		schema:       defaultHTTPSchema,
		version:      apiVersion1,
		apiURI:       relativeAPIath,
		repoURI:      relativeRepoPath,
		repoFilesURI: relativeRepoFilePath,
	}).init()
}

func (api *repoAPIURI) init() *repoAPIURI {
	api.baseRepoURI = fmt.Sprintf(`%s://%s:%d/%s/%s/%s`,
		api.schema, api.serverAddr, api.serverPort, api.apiURI, api.version, api.repoURI)
	api.baseRepoFilesURI = fmt.Sprintf(`%s://%s:%d/%s/%s/%s`,
		api.schema, api.serverAddr, api.serverPort, api.apiURI, api.version, api.repoFilesURI)
	return api
}

// PipyRepoClient Pipy Repo Client
type PipyRepoClient struct {
	apiURI           *repoAPIURI
	defaultTransport *http.Transport
	httpClient       *resty.Client
}

// NewRepoClient creates a Repo Client
func NewRepoClient(serverAddr string, serverPort uint16) *PipyRepoClient {
	return NewRepoClientWithTransport(
		serverAddr, serverPort,
		&http.Transport{
			DisableKeepAlives:  false,
			MaxIdleConns:       10,
			IdleConnTimeout:    60 * time.Second,
			DisableCompression: false,
		})
}

// NewRepoClientWithTransport creates a Repo Client with Transport
func NewRepoClientWithTransport(serverAddr string, serverPort uint16, transport *http.Transport) *PipyRepoClient {
	return NewRepoClientWithAPIBaseURLAndTransport(serverAddr, serverPort, transport)
}

// NewRepoClientWithAPIBaseURLAndTransport creates a Repo Client with ApiBaseUrl and Transport
func NewRepoClientWithAPIBaseURLAndTransport(serverAddr string, serverPort uint16, transport *http.Transport) *PipyRepoClient {
	repo := &PipyRepoClient{
		apiURI:           newRepoAPIURI(serverAddr, serverPort),
		defaultTransport: transport,
	}

	repo.httpClient = resty.New().
		SetTransport(repo.defaultTransport).
		SetScheme(repo.apiURI.schema).
		SetAllowGetMethodPayload(true).
		SetBaseURL(repo.apiURI.baseRepoURI).
		SetTimeout(5 * time.Second).
		SetDebug(false).
		EnableTrace()

	return repo
}

func (p *PipyRepoClient) isCodebaseExists(codebaseName string) (success bool, codebase *Codebase, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	resp, err = p.httpClient.R().
		SetResult(&Codebase{}).
		Get(codebaseName)

	if err == nil {
		success = true
		switch resp.StatusCode() {
		case http.StatusOK:
			codebase = resp.Result().(*Codebase)
			return
		default:
			return
		}
	}

	log.Err(err).Msgf("error happened while checking Codebase Exists[%s]", codebaseName)
	return
}

func (p *PipyRepoClient) getCodebase(codebaseName string) (success bool, codebase *Codebase, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	resp, err = p.httpClient.R().
		SetResult(&Codebase{}).
		Get(codebaseName)

	if err == nil {
		success = true
		switch resp.StatusCode() {
		case http.StatusOK:
			codebase = resp.Result().(*Codebase)
			return
		default:
			return
		}
	}

	log.Err(err).Msgf("error happened while getting Codebase[%s]", codebaseName)
	return
}

func (p *PipyRepoClient) createCodebase(version string, codebaseName string) (success bool, codebase *Codebase, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	resp, err = p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: version}).
		Post(codebaseName)

	if err == nil {
		switch resp.StatusCode() {
		case http.StatusOK, http.StatusCreated:
			return p.getCodebase(codebaseName)
		default:
			err = fmt.Errorf("error happened while creating Codebase[%s], status: %s reason:%s", codebaseName, resp.Status(), string(resp.Body()))
			return
		}
	}

	log.Err(err).Msgf("error happened while creating Codebase[%s]", codebaseName)
	return
}

func (p *PipyRepoClient) deriveCodebase(codebaseName, base string, version uint64) (success bool, codebase *Codebase, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	resp, err = p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: fmt.Sprintf("%d", version), Base: base}).
		Post(codebaseName)

	if err == nil {
		success = true
		switch resp.StatusCode() {
		case http.StatusOK, http.StatusCreated:
			success, codebase, err = p.getCodebase(codebaseName)
			return
		default:
			err = fmt.Errorf("error happened while deriving Codebase[%s] base[%s], reason: %s", codebaseName, base, resp.Status())
			return
		}
	}

	log.Err(err).Msgf("error happened while deriving Codebase[%s]", codebaseName)
	return
}

func (p *PipyRepoClient) upsertFile(path string, content interface{}) (success bool, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoFilesURI)
	// FIXME: temp solution, refine it later
	contentType := "text/plain"
	if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	}

	resp, err = p.httpClient.R().
		SetHeader("Content-Type", contentType).
		SetBody(content).
		Post(path)

	if err == nil {
		if resp.IsSuccess() {
			success = true
			return
		}
		err = fmt.Errorf("error happened while upserting file[%s], reason: %s", path, resp.Status())
		return
	}

	log.Err(err).Msgf("error happened while upserting file[%s]", path)
	return
}

// Delete codebase
func (p *PipyRepoClient) Delete(codebaseName string) (success bool, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	resp, err = p.httpClient.R().Delete(codebaseName)

	if err == nil {
		if resp.IsSuccess() {
			success = true
			return
		}
		err = fmt.Errorf("error happened while deleting codebase[%s], reason: %s", codebaseName, resp.Status())
		return
	}

	log.Err(err).Msgf("error happened while deleting codebase[%s]", codebaseName)
	return
}

// deleteFile delete codebase file
func (p *PipyRepoClient) deleteFile(fileName string) (success bool, err error) {
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoFilesURI)

	resp, err = p.httpClient.R().Delete(fileName)

	if err == nil {
		if resp.IsSuccess() {
			success = true
			return
		}
		err = fmt.Errorf("error happened while deleting codebase[%s], reason: %s", fileName, resp.Status())
		return
	}

	log.Err(err).Msgf("error happened while deleting codebase[%s]", fileName)
	return
}

// Commit the codebase, version is the current vesion of the codebase, it will be increased by 1 when committing
func (p *PipyRepoClient) commit(codebaseName string, version string) (success bool, err error) {
	var etag uint64
	var resp *resty.Response

	p.httpClient.SetBaseURL(p.apiURI.baseRepoURI)

	if etag, err = strconv.ParseUint(version, 10, 64); err != nil {
		return
	}

	resp, err = p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: fmt.Sprintf("%d", etag+1)}).
		SetResult(&Codebase{}).
		Patch(codebaseName)

	if err == nil {
		if resp.IsSuccess() {
			success = true
			return
		}
		err = fmt.Errorf("error happened while committing codebase[%s], reason: %s", codebaseName, resp.Status())
		return
	}

	log.Err(err).Msgf("error happened while committing codebase[%s]", codebaseName)
	return
}

// Batch submits multiple resources at once
func (p *PipyRepoClient) Batch(version string, batches []Batch) (success bool, err error) {
	if len(batches) == 0 {
		return
	}

	for _, batch := range batches {
		// 1. batch.Basepath, if not exists, create it
		log.Info().Msgf("batch.Basepath = %q", batch.Basepath)
		var codebaseV string
		var codebase *Codebase
		success, codebase, err = p.isCodebaseExists(batch.Basepath)
		if err != nil {
			return
		}
		if codebase != nil {
			// just getCodebase the version of codebase
			codebaseV = codebase.Version
		} else {
			log.Info().Msgf("%q doesn't exist in repo", batch.Basepath)
			success, codebase, err = p.createCodebase(version, batch.Basepath)

			if err != nil || !success || codebase == nil {
				log.Info().Msgf("Failure! Result = %#v", codebase)
				return
			}

			log.Info().Msgf("Success! Result = %#v", codebase)

			codebaseV = codebase.Version
		}

		// 2. upload each json to repo
		for _, item := range batch.Items {
			fullPath := fmt.Sprintf("%s%s/%s", batch.Basepath, item.Path, item.Filename)
			if item.Obsolete {
				log.Info().Msgf("Deleting config %q", fullPath)
				_, err = p.deleteFile(fullPath)
				if err != nil {
					log.Debug().Msgf("fail to delete %q", fullPath)
				}
			} else {
				log.Info().Msgf("Creating/updating config %q", fullPath)
				success, err = p.upsertFile(fullPath, item.Content)
				if err != nil || !success {
					return
				}
			}
		}

		// 3. commit the repo, so that changes can take effect
		log.Info().Msgf("Committing batch.Basepath = %q", batch.Basepath)
		if success, err = p.commit(batch.Basepath, codebaseV); err != nil || !success {
			return
		}
	}

	return
}

// DeriveCodebase derives Codebase
func (p *PipyRepoClient) DeriveCodebase(codebaseName, base string, version uint64) (success bool, err error) {
	var codebase *Codebase
	log.Info().Msgf("Checking if exists, codebase %q", codebaseName)
	success, codebase, err = p.isCodebaseExists(codebaseName)

	if err != nil {
		return
	}

	if codebase != nil {
		log.Info().Msgf("Codebase %q already exists, ignore deriving ...", codebaseName)
		return
	}

	log.Info().Msgf("Codebase %q doesn't exist, deriving ...", codebaseName)
	success, codebase, err = p.deriveCodebase(codebaseName, base, version)
	if err != nil {
		log.Err(err).Msgf("Deriving codebase %q", codebaseName)
		return
	}
	log.Info().Msgf("Successfully derived codebase %q", codebaseName)

	log.Info().Msgf("Committing the changes of codebase %q", codebaseName)
	if success, err = p.commit(codebaseName, codebase.Version); err != nil || !success {
		log.Err(err).Msgf("Committing codebase %q", codebaseName)
		return
	}

	log.Info().Msgf("Successfully committed codebase %q", codebaseName)
	return
}

// IsRepoUp checks whether the repo is up
func (p *PipyRepoClient) IsRepoUp() (success bool, err error) {
	if success, _, err = p.getCodebase("/"); err != nil || !success {
		log.Err(err).Msgf("Pipy Repo is not UP:")
		return
	}
	return
}
