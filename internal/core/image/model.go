package image

type ImageRef struct {
	Registry   string
	Repository string
	Reference  string
	Os         string
	Arch       string
}

type Descriptor struct {
	MediaType string
	Digest    string
	Size      int64
}

type Manifest struct {
	MediaType string
	Digest    string
	Config    Descriptor
	Layers    []Descriptor
}

type PullResult struct {
	ImageId    string
	Ref        string
	Manifest   string
	LayerCount int
}
