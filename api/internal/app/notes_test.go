package app

import "testing"

func TestValidateNoteNormalizesFields(t *testing.T) {
	t.Parallel()
	input := noteRequest{
		Title:    "title",
		Content:  "content",
		Tags:     []string{" Go ", "go", "数据库"},
		Category: " personal ",
		Status:   "TODO",
	}
	if err := validateNote(&input); err != nil {
		t.Fatal(err)
	}
	if input.Category != "personal" || input.Status != "todo" {
		t.Fatalf("unexpected normalized note: %+v", input)
	}
	if len(input.Tags) != 2 || input.Tags[0] != "Go" || input.Tags[1] != "数据库" {
		t.Fatalf("unexpected tags: %#v", input.Tags)
	}
}

func TestValidateNoteRejectsInvalidStatus(t *testing.T) {
	t.Parallel()
	input := noteRequest{Status: "unknown"}
	if err := validateNote(&input); err == nil {
		t.Fatal("expected invalid status to fail")
	}
}
