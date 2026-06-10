package replay

// Scene is a single replayable segment of a showboat document.
type Scene interface {
	kind() string
}

// TitleScene is the document header banner.
type TitleScene struct {
	Text string
}

func (TitleScene) kind() string { return "title" }

// NoteScene is commentary between code blocks.
type NoteScene struct {
	Lines      []string
	HasHeading bool
}

func (NoteScene) kind() string { return "note" }

// ExecScene is a command and its captured output.
type ExecScene struct {
	Lang    string
	Command string
	Output  string
}

func (ExecScene) kind() string { return "exec" }

// ImageScene is a captured image reference.
type ImageScene struct {
	Alt  string
	Path string
}

func (ImageScene) kind() string { return "image" }
