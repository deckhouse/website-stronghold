package extract

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
)

type BlobOpener func(digest string) (io.ReadCloser, error)

type imagesDigests struct {
	Stronghold string `json:"stronghold"`
}

func FindStrongholdDigest(layerDigests []string, open BlobOpener) (string, error) {
	const target = "images_digests.json"
	for _, digest := range layerDigests {
		data, err := findFileInLayer(digest, target, open)
		if err != nil {
			continue
		}
		var d imagesDigests
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		if d.Stronghold == "" {
			continue
		}
		return d.Stronghold, nil
	}
	return "", fmt.Errorf("images_digests.json not found in image layers")
}

func FindStrongholdBinary(layerDigests []string, open BlobOpener) ([]byte, error) {
	const target = "usr/bin/stronghold"
	for _, digest := range layerDigests {
		data, err := findFileInLayer(digest, target, open)
		if err != nil {
			continue
		}
		return data, nil
	}
	return nil, fmt.Errorf("usr/bin/stronghold not found in image layers")
}

func WriteStrongholdTar(w io.Writer, data []byte) error {
	tw := tar.NewWriter(w)
	hdr := &tar.Header{
		Name:   "stronghold",
		Mode:   0o755,
		Size:   int64(len(data)),
		Uid:    0,
		Gid:    0,
		Uname:  "root",
		Gname:  "root",
		Format: tar.FormatUSTAR,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(data); err != nil {
		return err
	}
	return tw.Close()
}

func ExtractStrongholdBinary(layerDigests []string, open BlobOpener, w io.Writer) error {
	data, err := FindStrongholdBinary(layerDigests, open)
	if err != nil {
		return err
	}
	return WriteStrongholdTar(w, data)
}

func findFileInLayer(digest, target string, open BlobOpener) ([]byte, error) {
	rc, err := open(digest)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	gz, err := gzip.NewReader(rc)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	target = normalizePath(target)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != 0 {
			continue
		}
		if normalizePath(hdr.Name) != target {
			continue
		}
		return io.ReadAll(tr)
	}
	return nil, fmt.Errorf("file %s not in layer", target)
}

func normalizePath(p string) string {
	p = strings.TrimPrefix(p, "./")
	return path.Clean(p)
}
