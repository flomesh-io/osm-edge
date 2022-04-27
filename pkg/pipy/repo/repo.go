package repo

import (
	"encoding/json"
	"fmt"
	"github.com/openservicemesh/osm/pkg/certificate"
	"github.com/openservicemesh/osm/pkg/errcode"
	"github.com/openservicemesh/osm/pkg/pipy"
	"github.com/openservicemesh/osm/pkg/pipy/registry"
	"github.com/openservicemesh/osm/pkg/utils"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	CertificateNoSerial   = "NoSerial"
	HttpStatusNotModified = 304
	HttpStatusNotFound    = 404
	RequestURIMaxLength   = 2083

	CodeBasePrefixRepo     = "/repo"
	CodebaseSuffixJs       = ".js"
	CodebaseSuffixJson     = ".json"
	CodebasePipyPolicyJson = "/pipy.json"
)

var (
	codebasePlugins = []string{
		"/main.js",
	}

	codebaseConfigs = []string{
		"/pipy.json",
	}

	codebaseLayout      = make(map[string]bool)
	codebasePluginsView = strings.Join(codebasePlugins, "\n")
	codebaseConfigsView = strings.Join(codebaseConfigs, "\n")

	pluginResources = make(map[pipy.RepoResource]*pipy.RepoResourceV)
)

func init() {
	for _, v := range codebaseConfigs {
		codebaseLayout[v] = true
	}

	hashCodes := make([]string, 0)
	for _, v := range codebasePlugins {
		codebaseLayout[v] = true
		resourceName := path.Join(CodeBasePrefixRepo, v)

		if resourceContent, err := loadCodebaseResource(resourceName); err == nil {
			resourceV := new(pipy.RepoResourceV)
			resourceV.Content = string(resourceContent)
			if hashCode, err := utils.HashFromString(resourceV.Content); err == nil {
				resourceV.Version = fmt.Sprintf("%d", hashCode)
				hashCodes = append(hashCodes, resourceV.Version)
			}
			pluginResources[pipy.RepoResource(v)] = resourceV
		}
	}
}

func (r *Repo) RegisterProxy(proxy *pipy.Proxy) (connectedProxy *ConnectedProxy, exists bool) {
	var actual interface{} = nil
	connectedProxy = &ConnectedProxy{
		proxy:       proxy,
		connectedAt: time.Now(),
	}
	actual, exists = r.connectedProxies.LoadOrStore(proxy.GetCertificateCommonName(), connectedProxy)
	if exists {
		connectedProxy = actual.(*ConnectedProxy)
	} else {
		proxyAddr := strings.Split(proxy.GetIP().String(), ":")[0]
		registry.AddCachedMeshPod(proxyAddr, proxy.GetCertificateCommonName().String())
		log.Debug().Str("proxy", proxy.String()).Msg("Registered new proxy")
	}
	return
}

func (r *Repo) UnregisterProxy(p *pipy.Proxy) {
	r.connectedProxies.Delete(p.GetCertificateCommonName())
	log.Debug().Msgf("Unregistered proxy %s", p.String())
}

func (r *Repo) GetPipyRepoHandler() http.Handler {

	r.server.proxyRegistry.SetReleaseCertificateCallback(func(podUID types.UID, endpointCN certificate.CommonName) {
		if actual, exist := r.connectedProxies.Load(endpointCN); exist {
			connectedProxy := actual.(*ConnectedProxy)
			close(connectedProxy.quit)
		}
	})

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		certCommonName, remoteAddr, ok := r.validateRequest(req)
		if !ok {
			return
		}

		connectedProxy, success := r.LoadOrRegisterProxy(certCommonName, remoteAddr)
		if !success {
			return
		}

		w.Header().Set(`server`, `pipy-repo`)
		w.Header().Set(`Connection`, `keep-alive`)

		if "POST" == req.Method {
			r.handlePipyReportRequest(connectedProxy, req)
			return
		}

		uri := req.RequestURI

		if strings.HasSuffix(uri, CodebaseSuffixJs) {
			r.handlePipyStaticScriptRequest(w, req, uri, certCommonName)
			return
		}

		if strings.HasSuffix(uri, CodebaseSuffixJson) && !strings.HasSuffix(uri, CodebasePipyPolicyJson) {
			r.handlePipyStaticJsonRequest(w, req, uri, certCommonName, connectedProxy)
			return
		}

		codebaseConf, latestRepoCodebaseV, codebaseReady := r.getCodebaseStatus(w, req, connectedProxy, certCommonName)
		if !codebaseReady {
			return
		}

		var pipyConf *PipyConf
		castOk := false
		if pipyConf, castOk = codebaseConf.(*PipyConf); !castOk {
			w.WriteHeader(HttpStatusNotFound)
			return
		}

		if strings.HasSuffix(uri, "/") {
			r.handlePipyCodebaseLayoutRequest(w, req, connectedProxy, pipyConf, latestRepoCodebaseV, certCommonName)
			return
		}

		if strings.HasSuffix(uri, CodebasePipyPolicyJson) {
			r.handlePipyPolicyJsonRequest(w, req, latestRepoCodebaseV, certCommonName, pipyConf)
			return
		}
	})
}

func (r *Repo) handlePipyPolicyJsonRequest(w http.ResponseWriter, req *http.Request, latestRepoCodebaseV string, certCommonName certificate.CommonName, pipyConf *PipyConf) {
	w.Header().Set(`Etag`, latestRepoCodebaseV)
	log.Trace().Str("Proxy", certCommonName.String()).Msgf("URI:%s RIP:%s ETag:%s",
		req.RequestURI, req.RemoteAddr, latestRepoCodebaseV)
	if "GET" == req.Method {
		if _, netErr := fmt.Fprint(w, string(pipyConf.bytes)); netErr != nil {
			log.Error().Err(netErr).Msgf("Error writing response content")
		}
	}
}

func (r *Repo) handlePipyCodebaseLayoutRequest(w http.ResponseWriter, req *http.Request,
	connectedProxy *ConnectedProxy, pipyConf *PipyConf, latestRepoCodebaseV string,
	certCommonName certificate.CommonName) {
	etag := RefreshPipyConf(connectedProxy.proxy, pipyConf)
	if len(etag) == 0 {
		etag = latestRepoCodebaseV
	}
	w.Header().Set(`Etag`, etag)
	log.Trace().Str("Proxy", certCommonName.String()).Msgf("URI:%s RIP:%s ETag:%s",
		req.RequestURI, req.RemoteAddr, latestRepoCodebaseV)
	if strings.EqualFold(`GET`, req.Method) {
		if _, netErr := fmt.Fprint(w, codebasePluginsView, "\n", codebaseConfigsView, "\n"); netErr != nil {
			log.Error().Err(netErr).Msgf("Error writing response content")
		}
	}
}

func (r *Repo) getCodebaseStatus(w http.ResponseWriter, req *http.Request, connectedProxy *ConnectedProxy,
	certCommonName certificate.CommonName) (interface{}, string, bool) {
	codebaseConf, latestRepoCodebaseV, codebaseReady := connectedProxy.proxy.GetCodebase()
	if !codebaseReady {
		newJob := func() *PipyConfGeneratorJob {
			return &PipyConfGeneratorJob{
				proxy:     connectedProxy.proxy,
				xdsServer: r.server,
				done:      make(chan struct{}),
			}
		}
		<-r.server.workqueues.AddJob(newJob())

		w.WriteHeader(HttpStatusNotModified)
		log.Debug().Str("Proxy", certCommonName.String()).Msgf("URI:%s RIP:%s Status:304",
			req.RequestURI, req.RemoteAddr)
		return nil, "", false
	}
	return codebaseConf, latestRepoCodebaseV, true
}

func (r *Repo) handlePipyStaticJsonRequest(w http.ResponseWriter, req *http.Request, uri string,
	certCommonName certificate.CommonName, connectedProxy *ConnectedProxy) {
	resourceName := strings.Replace(uri, fmt.Sprintf("/%s", certCommonName), "", 1)
	resourceURI := strings.TrimPrefix(resourceName, CodeBasePrefixRepo)
	if _, find := codebaseLayout[resourceURI]; find {
		resourceV, ok := connectedProxy.proxy.GetLatestRepoResources(pipy.RepoResource(resourceURI))
		if !ok {
			if resourceContent, loadErr := loadCodebaseResource(resourceName); loadErr == nil {
				resourceV = new(pipy.RepoResourceV)
				resourceV.Content = string(resourceContent)
				connectedProxy.proxy.SetLatestRepoResources(pipy.RepoResource(resourceURI), resourceV)
			}
		}

		if resourceV != nil {
			w.Header().Set(`Etag`, resourceV.Version)
			log.Trace().Str("Proxy", certCommonName.String()).Msgf("URI:%s RIP:%s ETag:%s",
				req.RequestURI, req.RemoteAddr, resourceV.Version)
			if "GET" == req.Method {
				if _, netErr := fmt.Fprint(w, resourceV.Content); netErr != nil {
					log.Error().Err(netErr).Msgf("Error writing response content")
				}
			}
		} else {
			w.WriteHeader(HttpStatusNotFound)
		}
	} else {
		w.WriteHeader(HttpStatusNotFound)
	}
}

func (r *Repo) handlePipyStaticScriptRequest(w http.ResponseWriter, req *http.Request, uri string,
	certCommonName certificate.CommonName) {
	resourceName := strings.Replace(uri, fmt.Sprintf("/%s", certCommonName), "", 1)
	resourceURI := strings.TrimPrefix(resourceName, CodeBasePrefixRepo)
	if _, find := codebaseLayout[resourceURI]; find {
		if resourceV, ok := pluginResources[pipy.RepoResource(resourceURI)]; ok {
			w.Header().Set(`Etag`, resourceV.Version)
			log.Trace().Str("Proxy", certCommonName.String()).Msgf("URI:%s RIP:%s ETag:%s",
				req.RequestURI, req.RemoteAddr, resourceV.Version)
			if "GET" == req.Method {
				if _, netErr := fmt.Fprint(w, resourceV.Content); netErr != nil {
					log.Error().Err(netErr).Msgf("Error writing response content")
				}
			}
		} else {
			w.WriteHeader(HttpStatusNotFound)
		}
	} else {
		w.WriteHeader(HttpStatusNotFound)
	}
}

func (r *Repo) handlePipyReportRequest(connectedProxy *ConnectedProxy, req *http.Request) {
	connectedProxy.lastReportAt = time.Now()
	if bytes, netErr := ioutil.ReadAll(req.Body); netErr == nil {
		pjsReport := new(PipyReport)
		if netErr = json.Unmarshal(bytes, pjsReport); netErr == nil {
			connectedProxy.proxy.SetReportRepoCodebaseV(pjsReport.Version)
		}
	}
}

func (r *Repo) LoadOrRegisterProxy(certCommonName certificate.CommonName, remoteAddr *net.TCPAddr) (*ConnectedProxy, bool) {
	proxy, err := pipy.NewProxy(certCommonName, CertificateNoSerial, remoteAddr)
	if err != nil {
		log.Error().Err(err).Str(errcode.Kind, errcode.GetErrCodeWithMetric(errcode.ErrInitializingProxy)).
			Msgf("Error initializing proxy with certificate Common Name=%s", certCommonName)
		return nil, false
	}

	connectedProxy, exists := r.RegisterProxy(proxy)
	if !exists {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			if aggregatedErr := r.server.InformTrafficPolicies(r, wg, connectedProxy); aggregatedErr != nil {
				wg.Done()
				log.Error().Err(aggregatedErr).Msgf("Pipy Aggregated Traffic Policies Error.")
			}
		}()
		wg.Wait()

		if connectedProxy.initError != nil {
			log.Error().Err(connectedProxy.initError).Msgf("Pipy Aggregated Traffic Policies Error.")
			return nil, false
		}
	}
	return connectedProxy, true
}

func (r *Repo) validateRequest(req *http.Request) (certificate.CommonName, *net.TCPAddr, bool) {
	if len(req.RequestURI) > RequestURIMaxLength {
		return "", nil, false
	}

	pathVars := strings.SplitN(req.RequestURI, "/", 4)
	if len(pathVars) < 4 || "repo" != pathVars[1] {
		return "", nil, false
	}

	remoteAddr, err := net.ResolveTCPAddr("tcp", req.RemoteAddr)
	if err != nil {
		log.Error().Err(err).Msgf(err.Error())
		return "", nil, false
	}

	return certificate.CommonName(pathVars[2]), remoteAddr, true
}

func loadCodebaseResource(resourceName string) (resourceContent []byte, err error) {
	var resourceFile *os.File
	resourceFile, err = os.Open(resourceName)
	if err != nil {
		log.Error().Err(err).Msgf("Error opening resource file[%s]", resourceName)
		return
	}

	defer func(resourceFile *os.File) {
		_ = resourceFile.Close()
	}(resourceFile)

	resourceContent, err = ioutil.ReadAll(resourceFile)
	if err != nil {
		log.Error().Err(err).Msgf("Error reading resource file[%s]", resourceName)
		return
	}
	return
}
