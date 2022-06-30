package client

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// PipyRepoClient Pipy Repo Client
type PipyRepoClient struct {
	baseURL          string
	defaultTransport *http.Transport
	httpClient       *resty.Client
}

// NewRepoClient creates a Repo Client
func NewRepoClient(repoRootAddr string) *PipyRepoClient {
	return NewRepoClientWithTransport(
		repoRootAddr,
		&http.Transport{
			DisableKeepAlives:  false,
			MaxIdleConns:       10,
			IdleConnTimeout:    60 * time.Second,
			DisableCompression: false,
		})
}

// NewRepoClientWithTransport creates a Repo Client with Transport
func NewRepoClientWithTransport(repoRootAddr string, transport *http.Transport) *PipyRepoClient {
	return NewRepoClientWithAPIBaseURLAndTransport(
		fmt.Sprintf(pipyRepoAPIBaseURLTemplate, defaultHTTPSchema, repoRootAddr),
		transport,
	)
}

// NewRepoClientWithAPIBaseURLAndTransport creates a Repo Client with ApiBaseUrl and Transport
func NewRepoClientWithAPIBaseURLAndTransport(repoAPIBaseURL string, transport *http.Transport) *PipyRepoClient {
	repo := &PipyRepoClient{
		baseURL:          repoAPIBaseURL,
		defaultTransport: transport,
	}

	repo.httpClient = resty.New().
		SetTransport(repo.defaultTransport).
		SetScheme(defaultHTTPSchema).
		SetAllowGetMethodPayload(true).
		SetBaseURL(repo.baseURL).
		SetTimeout(5 * time.Second).
		SetDebug(false).
		EnableTrace()

	return repo
}

func (p *PipyRepoClient) isCodebaseExists(path string) (bool, *Codebase) {
	resp, err := p.httpClient.R().
		SetResult(&Codebase{}).
		Get(path)

	if err == nil {
		switch resp.StatusCode() {
		case http.StatusNotFound:
			return false, nil
		case http.StatusOK:
			return true, resp.Result().(*Codebase)
		}
	}

	log.Err(err).Msgf("error happened while getting path %q", path)
	return false, nil
}

func (p *PipyRepoClient) get(path string) (*Codebase, error) {
	resp, err := p.httpClient.R().
		SetResult(&Codebase{}).
		Get(path)

	if err != nil {
		log.Err(err).Msgf("Failed to get path %q", path)
		return nil, err
	}

	return resp.Result().(*Codebase), nil
}

func (p *PipyRepoClient) createCodebase(path string) (*Codebase, error) {
	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: 1}).
		Post(path)

	if err != nil {
		log.Err(err).Msgf("failed to create codebase %q", path)
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create codebase %q, reason: %s", path, resp.Status())
	}

	codebase, err := p.get(path)
	if err != nil {
		return nil, err
	}

	return codebase, nil
}

func (p *PipyRepoClient) deriveCodebase(path, base string) (*Codebase, error) {
	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: 1, Base: base}).
		Post(path)

	if err != nil {
		log.Err(err).Msgf("Failed to derive codebase codebase: path: %q, base: %q", path, base)
		return nil, err
	}

	switch resp.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		log.Info().Msgf("Status code is %d, stands for success.", resp.StatusCode())
	default:
		log.Error().Msgf("Response contains error: %#v", resp.Status())
		return nil, fmt.Errorf("failed to derive codebase codebase: path: %q, base: %q, reason: %s", path, base, resp.Status())
	}

	log.Info().Msgf("Getting info of codebase %q", path)
	codebase, err := p.get(path)
	if err != nil {
		log.Error().Msgf("Failed to get info of codebase %q", path)
		return nil, err
	}

	log.Info().Msgf("Successfully derived codebase: %#v", codebase)
	return codebase, nil
}

func (p *PipyRepoClient) upsertFile(path string, content interface{}) error {
	// FIXME: temp solution, refine it later
	contentType := "text/plain"
	if strings.HasSuffix(path, ".json") {
		contentType = "application/json"
	}

	resp, err := p.httpClient.R().
		SetHeader("Content-Type", contentType).
		SetBody(content).
		Post(path)

	if err != nil {
		log.Err(err).Msgf("error happened while trying to upsert %q to repo", path)
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	errMsg := "repo server responsed with error HTTP code: %d, error: %s"
	log.Error().Msgf(errMsg, resp.StatusCode(), resp.Status())
	return fmt.Errorf(errMsg, resp.StatusCode(), resp.Status())
}

// Delete codebase
func (p *PipyRepoClient) Delete(path string) {
	// DELETE, as pipy repo doesn't support deletion yet, this's not implemented
	panic("implement me")
}

// Commit the codebase, version is the current vesion of the codebase, it will be increased by 1 when committing
func (p *PipyRepoClient) commit(path string, version int64) error {
	resp, err := p.httpClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(Codebase{Version: version + 1}).
		SetResult(&Codebase{}).
		Post(path)

	if err != nil {
		return err
	}

	if resp.IsSuccess() {
		return nil
	}

	err = fmt.Errorf("failed to commit codebase %q, reason: %s", path, resp.Status())
	log.Err(err)

	return err
}

// Batch submits multiple resources at once
func (p *PipyRepoClient) Batch(batches []Batch) error {
	if len(batches) == 0 {
		return nil
	}

	for _, batch := range batches {
		// 1. batch.Basepath, if not exists, create it
		log.Info().Msgf("batch.Basepath = %q", batch.Basepath)
		var codebaseV int64
		exists, codebase := p.isCodebaseExists(batch.Basepath)
		if exists {
			// just get the version of codebase
			codebaseV = codebase.Version
		} else {
			log.Info().Msgf("%q doesn't exist in repo", batch.Basepath)
			result, err := p.createCodebase(batch.Basepath)
			if err != nil {
				log.Err(err).Msgf("Not able to create the codebase %q", batch.Basepath)
				return err
			}

			log.Info().Msgf("Result = %#v", result)

			codebaseV = result.Version
		}

		// 2. upload each json to repo
		for _, item := range batch.Items {
			fullPath := fmt.Sprintf("%s%s/%s", batch.Basepath, item.Path, item.Filename)
			log.Info().Msgf("Creating/updating config %q", fullPath)
			err := p.upsertFile(fullPath, item.Content)
			if err != nil {
				log.Err(err).Msgf("Upsert %q error", fullPath)
				return err
			}
		}

		// 3. commit the repo, so that changes can take effect
		log.Info().Msgf("Committing batch.Basepath = %q", batch.Basepath)
		// NOT a valid version, ignore committing
		if codebaseV == -1 {
			err := fmt.Errorf("%d is not a valid version", codebaseV)
			log.Err(err)
			return err
		}
		if err := p.commit(batch.Basepath, codebaseV); err != nil {
			log.Err(err).Msgf("Error happened while committing the codebase %q", batch.Basepath)
			return err
		}
	}

	return nil
}

// DeriveCodebase derives Codebase
func (p *PipyRepoClient) DeriveCodebase(path, base string) error {
	log.Info().Msgf("Checking if exists, codebase %q", path)
	exists, _ := p.isCodebaseExists(path)

	if exists {
		log.Info().Msgf("Codebase %q already exists, ignore deriving ...", path)
	} else {
		log.Info().Msgf("Codebase %q doesn't exist, deriving ...", path)
		result, err := p.deriveCodebase(path, base)
		if err != nil {
			log.Err(err).Msgf("Deriving codebase %q", path)
			return err
		}
		log.Info().Msgf("Successfully derived codebase %q", path)

		log.Info().Msgf("Committing the changes of codebase %q", path)
		if err = p.commit(path, result.Version); err != nil {
			log.Err(err).Msgf("Committing codebase %q", path)
			return err
		}
		log.Info().Msgf("Successfully committed codebase %q", path)
	}

	return nil
}

// IsRepoUp checks whether the repo is up
func (p *PipyRepoClient) IsRepoUp() bool {
	_, err := p.get("/")
	return err == nil
}
