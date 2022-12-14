package repo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/openservicemesh/osm/pkg/announcements"
	"github.com/openservicemesh/osm/pkg/sidecar/providers/pipy/client"
)

func (s *Server) pluginListener() {
	pluginUpdatePubSub := s.msgBroker.GetPluginUpdatePubSub()
	pluginUpdateChan := pluginUpdatePubSub.Sub(announcements.PluginUpdate.String())
	defer s.msgBroker.Unsub(pluginUpdatePubSub, pluginUpdateChan)

	// Wait for two informer synchronization periods
	slidingTimer := time.NewTimer(time.Second * 10)
	defer slidingTimer.Stop()

	reconfirm := true

	for {
		select {
		case <-pluginUpdateChan:
			// Wait for an informer synchronization period
			slidingTimer.Reset(time.Second * 5)
			// Avoid data omission
			reconfirm = true

		case <-slidingTimer.C:
			if s.updatePlugins() {
				s.fireExistProxies()
			}
			if reconfirm {
				reconfirm = false
				slidingTimer.Reset(time.Second * 10)
			}
		}
	}
}

func (s *Server) updatePlugins() bool {
	var pluginItems []client.BatchItem
	var pluginVers []string
	newPluginSet := mapset.NewSet()

	plugins := s.catalog.GetPlugins()
	for _, plugin := range plugins {
		uri := plugin.GetPluginURI()
		bytes := []byte(plugin.Script)
		newPluginSet.Add(uri)
		pluginItems = append(pluginItems, client.BatchItem{
			Filename: uri,
			Content:  bytes,
		})
		pluginVers = append(pluginVers, fmt.Sprintf("%s:%d", uri, hash(bytes)))
	}

	diffSet := s.pluginSet.Difference(newPluginSet)
	diffPluginURIs := diffSet.ToSlice()
	for _, pluginURI := range diffPluginURIs {
		pluginItems = append(pluginItems, client.BatchItem{
			Filename: pluginURI.(string),
			Obsolete: true,
		})
	}
	if len(pluginItems) > 0 {
		sort.Strings(pluginVers)
		pluginSetHash := hash([]byte(strings.Join(pluginVers, "")))
		pluginSetVersion := fmt.Sprintf("%d", pluginSetHash)
		_, err := s.repoClient.Batch(pluginSetVersion, []client.Batch{
			{
				Basepath: osmCodebase,
				Items:    pluginItems,
			},
		})
		if err != nil {
			log.Error().Err(err)
		} else {
			s.pluginSet = newPluginSet
			s.pluginSetVersion = pluginSetVersion
		}
		return true
	}

	return false
}
