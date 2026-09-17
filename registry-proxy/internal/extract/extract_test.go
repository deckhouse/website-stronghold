package extract

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func TestFindStrongholdDigest(t *testing.T) {
	layers := map[string][]byte{
		"sha256:layer1": tarGzFile(t, "other.txt", []byte("x")),
		"sha256:layer2": tarGzFile(t, "images_digests.json", []byte(`{"stronghold":"sha256:abc"}`)),
	}

	digest, err := FindStrongholdDigest([]string{"sha256:layer1", "sha256:layer2"}, fakeOpener(layers))
	if err != nil {
		t.Fatal(err)
	}
	if digest != "sha256:abc" {
		t.Fatalf("got %q", digest)
	}
}

func TestWriteStrongholdTar(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteStrongholdTar(&buf, []byte("binary")); err != nil {
		t.Fatal(err)
	}

	tr := tar.NewReader(&buf)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Name != "stronghold" || hdr.Mode != 0o755 || hdr.Uid != 0 || hdr.Gid != 0 {
		t.Fatalf("unexpected header: %+v", hdr)
	}
	data, err := io.ReadAll(tr)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "binary" {
		t.Fatalf("got %q", data)
	}
}

func tarGzFile(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var raw bytes.Buffer
	tw := tar.NewWriter(&raw)
	if err := tw.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	if _, err := zw.Write(raw.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return gz.Bytes()
}

func fakeOpener(layers map[string][]byte) BlobOpener {
	return func(digest string) (io.ReadCloser, error) {
		data, ok := layers[digest]
		if !ok {
			return io.NopCloser(bytes.NewReader(nil)), io.EOF
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}
