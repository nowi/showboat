package replay

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func updateGolden() bool {
	return os.Getenv("SHOWBOAT_UPDATE_GOLDEN") != ""
}

func TestParseDocumentBasic(t *testing.T) {
	scenes := mustParse(t, "basic.md")
	if len(scenes) != 2 {
		t.Fatalf("got %d scenes, want 2", len(scenes))
	}
	title, ok := scenes[0].(TitleScene)
	if !ok || title.Text != "Hello Replay" {
		t.Fatalf("unexpected title scene: %#v", scenes[0])
	}
	exec, ok := scenes[1].(ExecScene)
	if !ok || exec.Command != "echo hello" || exec.Output != "hello\n" {
		t.Fatalf("unexpected exec scene: %#v", scenes[1])
	}
}

func TestParseDocumentMultiline(t *testing.T) {
	scenes := mustParse(t, "multiline.md")
	exec, ok := scenes[1].(ExecScene)
	if !ok {
		t.Fatalf("expected exec scene, got %#v", scenes[1])
	}
	if exec.Command != "echo line1\necho line2" {
		t.Fatalf("command = %q", exec.Command)
	}
}

func TestGoldenCast(t *testing.T) {
	for _, name := range []string{"basic", "multiline"} {
		t.Run(name, func(t *testing.T) {
			mdPath := filepath.Join("testdata", name+".md")
			castPath := filepath.Join("testdata", name+".cast")

			f, err := os.Open(mdPath)
			if err != nil {
				t.Fatal(err)
			}
			scenes, err := ParseDocument(f)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}

			var buf bytes.Buffer
			opts := DefaultOptions()
			emitter := NewEmitter(&buf, opts)
			docTitle := ""
			for _, s := range scenes {
				if ts, ok := s.(TitleScene); ok {
					docTitle = ts.Text
					break
				}
			}
			if err := emitter.RenderScenes(scenes, docTitle); err != nil {
				t.Fatal(err)
			}
			got := buf.Bytes()

			if updateGolden() {
				if err := os.WriteFile(castPath, got, 0644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(castPath)
			if err != nil {
				t.Fatalf("missing golden %s (run SHOWBOAT_UPDATE_GOLDEN=1 go test ./replay/...): %v", castPath, err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("cast output differs from golden %s", castPath)
			}
		})
	}
}

func TestReplayVerify(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "demo.md")
	cast := filepath.Join(dir, "demo.cast")

	content := "# Verify Test\n\n*2026-06-10T12:00:00Z*\n\n```bash\necho ok\n```\n\n```output\nok\n```\n"
	if err := os.WriteFile(md, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions()
	opts.Output = cast
	if err := Replay(md, opts); err != nil {
		t.Fatal(err)
	}

	opts.Verify = true
	if err := Replay(md, opts); err != nil {
		t.Fatalf("verify should pass: %v", err)
	}

	if err := os.WriteFile(cast, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}
	opts.Verify = true
	if err := Replay(md, opts); err == nil {
		t.Fatal("verify should fail on drift")
	}
}

func mustParse(t *testing.T, name string) []Scene {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	scenes, err := ParseDocument(f)
	if err != nil {
		t.Fatal(err)
	}
	return scenes
}
