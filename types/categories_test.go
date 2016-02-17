package types

import "testing"

func TestGen(t *testing.T) {
	if TV_Anime.String() != "TV_Anime" {
		t.Fatalf("Expected string rep of TV_Anime to be 'TV_Anime', got %s", TV_Anime.String())
	}
	unk := Category(-1).String()
	if unk != "Unknown" {
		t.Fatalf("Expected string rep of -1 to be 'Unknown', got %s", unk)
	}
}
