package conf

import (
    "io"
)

type SyntaxHandler interface {
    Read(rd io.Reader) (interface{}, error)
    Load(filename string) (interface{}, error)
    Parse(b []byte) (interface{}, error)
    LoadDir(dirname string) (interface{}, error)
    Write(wr io.Writer, data interface{}) (int, error)
    Dump(data interface{}) ([]byte, error)
}

