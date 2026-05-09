package archive

import "testing"

func TestDetect(t *testing.T) {
	tests := []struct {
		name    string
		want    Format
		wantErr bool
	}{
		{"archive.tar.zst", Zstd, false},
		{"archive.tar.gz", Gzip, false},
		{"archive.tgz", Gzip, false},
		{"ARCHIVE.TAR.ZST", Zstd, false},
		{"ARCHIVE.TAR.GZ", Gzip, false},
		{"archive.zip", 0, true},
		{"archive.tar", Tar, false},
		{"archive.rar", 0, true},
		{"noextension", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detect(tt.name)
			if (err != nil) != tt.wantErr {
				t.Fatalf("detect(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("detect(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFormatString(t *testing.T) {
	if Zstd.String() != ".tar.zst" {
		t.Fatalf("Zstd.String() = %q, want %q", Zstd.String(), ".tar.zst")
	}
	if Gzip.String() != ".tar.gz" {
		t.Fatalf("Gzip.String() = %q, want %q", Gzip.String(), ".tar.gz")
	}
}
