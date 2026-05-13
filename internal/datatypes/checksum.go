package datatypes

type ChecksumAlgo string

const ChecksumAlgoSHA256 ChecksumAlgo = "SHA-256"

type Checksum struct {
	Algorithm ChecksumAlgo
	Hash      string
}
