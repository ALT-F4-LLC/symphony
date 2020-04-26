package block

// Block : block node
type Block struct {
	Flags *flags
}

// New : creates new block node
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	b := &Block{
		Flags: flags,
	}

	return b, nil
}
