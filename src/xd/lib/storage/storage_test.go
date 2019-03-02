package storage

import (
	"crypto/rand"
	"io"
	"testing"
	"xd/lib/common"
	"xd/lib/fs"
	"xd/lib/metainfo"
	"xd/lib/mktorrent"
)

const testPieceLen = 65536

func createRandomTorrent(testFname string) (*metainfo.TorrentFile, error) {
	f, err := fs.STD.OpenFileWriteOnly(testFname)
	if err != nil {
		return nil, err
	}
	_, err = io.CopyN(f, rand.Reader, (testPieceLen*8)+128)
	f.Sync()
	f.Close()

	return mktorrent.MakeTorrent(fs.STD, testFname, testPieceLen)
}

func TestStorage(t *testing.T) {

	st := &FsStorage{
		MetaDir:    "storage",
		DataDir:    "data",
		SeedingDir: "seeding",
		FS:         fs.STD,
	}

	err := st.Init()
	if err != nil {
		t.Log("failed to init storage")
		t.Fail()
		return
	}
	fname := st.FS.Join(st.DataDir, "test.bin")
	meta, err := createRandomTorrent(fname)
	if err != nil {
		t.Logf("failed to make torrent: %s", err.Error())
		t.Fail()
		return
	}

	torrent, err := st.OpenTorrent(meta)
	if err != nil {
		t.Log("failed to open torrent")
		t.Fail()
		return
	}
	err = torrent.VerifyAll()
	if err != nil {
		t.Log("verify all failed")
		t.Fail()
		return
	}
	var pc common.PieceData
	err = torrent.GetPiece(common.PieceRequest{
		Index:  0,
		Begin:  0,
		Length: 16384,
	}, &pc)

	if err != nil {
		t.Fail()
		return
	}

	err = torrent.PutChunk(&pc)
	if err != nil {
		t.Fail()
		return
	}

	err = torrent.VerifyPiece(0)
	if err != nil {
		t.Fail()
		return
	}

}
