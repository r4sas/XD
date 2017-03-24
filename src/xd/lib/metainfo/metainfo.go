package metainfo

import (
	"bytes"
	"crypto/sha1"
	"github.com/zeebo/bencode"
	"io"
	"os"
	"path/filepath"
	"xd/lib/common"
	"xd/lib/log"
)

type FilePath []string

// get filepath
func (f FilePath) FilePath() string {
	return filepath.Join(f...)
}

/** open file using base path */
func (f FilePath) Open(base string) (*os.File, error) {
	return os.OpenFile(filepath.Join(base, f.FilePath()), os.O_RDWR|os.O_CREATE, 0600)
}

type FileInfo struct {
	// length of file
	Length uint64 `bencode:"length"`
	// relative path of file
	Path FilePath `bencode:"path"`
	// md5sum
	Sum []byte `bencode:"md5sum,omitempty"`
}

// info section of torrent file
type Info struct {
	// length of pices in bytes
	PieceLength uint32 `bencode:"piece length"`
	// piece data
	Pieces []byte `bencode:"pieces"`
	// name of root file
	Path string `bencode:"name"`
	// file metadata
	Files []FileInfo `bencode:"files,omitempty"`
	// private torrent
	Private int64 `bencode:"private,omitempty"`
	// length of file in signle file mode
	Length uint64 `bencode:"length,omitempty"`
	// md5sum
	Sum []byte `bencode:"md5sum,omitempty"`
}

// get fileinfos from this info section
func (i Info) GetFiles() (infos []FileInfo) {
	if i.Length > 0 {
		infos = append(infos, FileInfo{
			Length: i.Length,
			Path:   FilePath([]string{i.Path}),
			Sum:    i.Sum,
		})
	} else {
		infos = append(infos, i.Files...)
	}
	return
}

// check if a piece is valid against the pieces in this info section
func (i Info) CheckPiece(p *common.PieceData) bool {
	idx := p.Index * 20
	if i.NumPieces() > p.Index {
		log.Debugf("sum len=%d idx=%d ih=%d", len(p.Data), idx, len(i.Pieces))
		h := sha1.Sum(p.Data)
		expected := i.Pieces[idx : idx+20]
		return bytes.Equal(h[:], expected)
	}
	log.Error("piece index out of bounds")
	return false
}

func (i Info) NumPieces() uint32 {
	return uint32(len(i.Pieces) / 20)
}

// a torrent file
type TorrentFile struct {
	Info         Info       `bencode:"info"`
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Created      int64      `bencode:"created"`
	Comment      []byte     `bencode:"comment"`
	CreatedBy    []byte     `bencode:"created by"`
	Encoding     []byte     `bencode:"encoding"`
}

// get total size of files from torrent info section
func (tf *TorrentFile) TotalSize() uint64 {
	if tf.IsSingleFile() {
		return tf.Info.Length
	}
	total := uint64(0)
	for _, f := range tf.Info.Files {
		total += f.Length
	}
	return total
}

func (tf *TorrentFile) GetAllAnnounceURLS() (l []string) {
	if len(tf.Announce) > 0 {
		l = append(l, tf.Announce)
	}
	for _, al := range tf.AnnounceList {
		for _, a := range al {
			if len(a) > 0 {
				l = append(l, a)
			}
		}
	}
	return
}

func (tf *TorrentFile) TorrentName() string {
	return string(tf.Info.Path)
}

// calculate infohash
func (tf *TorrentFile) Infohash() (ih common.Infohash) {
	s := sha1.New()
	enc := bencode.NewEncoder(s)
	enc.Encode(&tf.Info)
	d := s.Sum(nil)
	copy(ih[:], d[:])
	return
}

// return true if this torrent is for a single file
func (tf *TorrentFile) IsSingleFile() bool {
	return tf.Info.Length > 0
}

// bencode this file via an io.Writer
func (tf *TorrentFile) BEncode(w io.Writer) (err error) {
	enc := bencode.NewEncoder(w)
	err = enc.Encode(tf)
	return
}

// load from an io.Reader
func (tf *TorrentFile) BDecode(r io.Reader) (err error) {
	dec := bencode.NewDecoder(r)
	err = dec.Decode(tf)
	return
}
