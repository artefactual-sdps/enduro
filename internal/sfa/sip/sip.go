package sip

import (
	"io/fs"
	"os"
	"path/filepath"
)

type SFASip struct {
	Path            string
	Header          *Header
	Content         *Content
	XSDPresent      bool
	MetadataPresent bool
	Unexpected      []string
	Files           []string
}

type Header struct {
	Path string
}
type Content struct {
	Path string
}

func NewSFASip(path string) (*SFASip, error) {
	s := &SFASip{Path: path}
	nodes, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		nodePath := filepath.Join(path, n.Name())
		if !n.IsDir() {
			s.Unexpected = append(s.Unexpected, nodePath)
		} else {
			switch n.Name() {
			case "header":
				s.Header = &Header{Path: nodePath}
				headerNodes, err := os.ReadDir(nodePath)
				if err != nil {
					return nil, err
				}
				for _, n := range headerNodes {
					if n.IsDir() {
						if n.Name() == "xsd" {
							s.XSDPresent = true
						}
					} else if n.Name() == "metadata.xml" {
						s.MetadataPresent = true
					} else {
						s.Unexpected = append(s.Unexpected, nodePath)
					}
				}
			case "content":
				s.Content = &Content{Path: nodePath}
				s.Files, err = initFiles(nodePath)
				if err != nil {
					return nil, err
				}
			default:
				s.Unexpected = append(s.Unexpected, nodePath)
			}
		}
	}

	return s, nil
}

func initFiles(path string) ([]string, error) {
	res := []string{}
	err := filepath.WalkDir(path,
		func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			res = append(res, p)
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return res, nil
}
