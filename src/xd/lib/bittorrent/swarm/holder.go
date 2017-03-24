package swarm

import (
	"sync"
	"xd/lib/common"
	"xd/lib/storage"
)

// torrent swarm container
type Holder struct {
	sw       *Swarm
	st       storage.Storage
	access   sync.RWMutex
	torrents map[string]*Torrent
}

func (h *Holder) addTorrent(t storage.Torrent) {
	h.access.Lock()
	defer h.access.Unlock()
	tr := newTorrent(t)
	h.torrents[t.Infohash().Hex()] = tr
	go h.sw.startTorrent(tr)
}

func (h *Holder) ForEachTorrent(visit func(*Torrent)) {
	var torrents []*Torrent
	h.access.Lock()
	for _, t := range h.torrents {
		torrents = append(torrents, t)
	}
	h.access.Unlock()
	for _, t := range torrents {
		visit(t)
	}
}

// find a torrent by infohash
// returns nil if we don't have a torrent with this infohash
func (h *Holder) GetTorrent(ih common.Infohash) (t *Torrent) {
	h.access.Lock()
	t, _ = h.torrents[ih.Hex()]
	h.access.Unlock()
	return
}
