package conf

import (
    "os"
    "io"
    "io/ioutil"
    "bytes"
    "bufio"
    "strings"
    "regexp"
    "path/filepath"
)

type SyntaxHandlerBD struct {
}

var (
    bdCommentString = map[string]bool{
        "#": true,
        ";": true,
    }
    bdValidKeyPattern, _ = regexp.Compile("^[A-Za-z][A-Za-z0-9-_]*$")
)

func (h *SyntaxHandlerBD) Read(rd io.Reader) (data interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    var sc = bufio.NewScanner(rd)
    var l string;
    var s string;
    var g []string;
    var stack = make([]map[string]interface{}, 0, 512)
    var curr = data.(map[string]interface{})
    var ok bool
    var lg int
    var lv int
    var k string
    var v interface{}
    var fp *os.File

    var lineno = 0
    var filename = "<io.Reader>"

    fp, ok = rd.(*os.File)
    if ok {
        filename = fp.Name()
    }

    stack = append(stack, curr)
    for sc.Scan() {
        lineno ++
        l = sc.Text()
        l = strings.TrimSpace(l)
        if len(l) == 0 {
            continue
        }

        // comment
        if _, ok = bdCommentString[string(l[0])]; ok {
            continue
        }

        // for group
        if l[0] == '[' && l[len(l) - 1] == ']' {
            s = strings.Trim(l, "[]")
            g = strings.Split(s, ".")
            curr = data.(map[string]interface{})
            lg = len(g)
            for lv, k = range g {
                if k == "" {
                    // we get . here
                    curr = stack[lv + 1]
                } else {
                    // . but with a string key
                    if _, ok = curr[k]; ok {
                        curr = curr[k].(map[string]interface{})
                        stack = stack[0:lv + 1]
                        stack = append(stack, curr)
                    }
                }
            }
            stack = stack[0:lg]
            k = g[lg - 1]
            v = make(map[string]interface{})
            // check tail key
            if k[0] == '@' {
                // array here
                k = string(k[1:])
                if ! bdValidKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ! ok {
                    curr[k] = make([]interface{}, 0, 0)
                }
                curr[k] = append(curr[k].([]interface{}), v)
            } else {
                // map here
                if ! bdValidKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ok {
                    err = &ConfError{"duplicate key", filename, lineno}
                    return
                }
                curr[k] = v
            }
            curr = v.(map[string]interface{})
            stack = append(stack, curr)
        } else {
            g = strings.SplitN(l, ":", 2)
            if len(g) != 2 {
                err = &ConfError{"syntax error", filename, lineno}
                return
            }
            k = strings.TrimSpace(g[0])
            v = strings.TrimSpace(g[1])
            if k[0] == '@' {
                k = string(k[1:])
                if ! bdValidKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                if _, ok = curr[k]; ! ok {
                    curr[k] = make([]interface{}, 0, 0)
                }
                curr[k] = append(curr[k].([]interface{}), v)
            } else {
                if ! bdValidKeyPattern.Match([]byte(k)) {
                    err = &ConfError{"invalid key", filename, lineno}
                    return
                }
                curr[k] = v
            }
        }
    }
    return
}

func (h *SyntaxHandlerBD) Parse(b []byte) (data interface{}, err error) {
    var buf = bytes.NewBuffer(b)
    return h.Read(buf)
}

func (h *SyntaxHandlerBD) Load(filename string) (data interface{}, err error) {
    var rd *os.File
    rd, err = os.OpenFile(filename, os.O_RDONLY, 0644)
    if err != nil {
        return
    }
    defer rd.Close()
    data, err = h.Read(rd)
    return
}

func (h *SyntaxHandlerBD) LoadDir(dir string) (data interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    var dirs = make([]os.FileInfo, 0, 512)
    var files = make([]os.FileInfo, 0, 512)
    var dirContent []os.FileInfo
    var hasGlobal = false
    var f os.FileInfo
    var k string
    var ok bool

    dirContent, err = ioutil.ReadDir(dir)
    if err != nil {
        return
    }
    for _, f = range dirContent {
        if f.Name()[0] == '.' {
            continue
        } else if f.IsDir() {
            dirs = append(dirs, f)
        } else if f.Name() == "_global.conf" {
            hasGlobal = true
        } else if filepath.Ext(f.Name()) == ".conf" {
            files = append(files, f)
        }
    }
    if hasGlobal {
        data, err = h.Load(filepath.Join(dir, "_global.conf"))
        if err != nil {
            return
        }
    }
    for _, f = range files {
        k = strings.TrimSuffix(f.Name(), ".conf")
        if _, ok = data.(map[string]interface{})[k]; ! ok {
            data.(map[string]interface{})[k], err = h.Load(filepath.Join(dir, f.Name()))
            if err != nil {
                return
            }
        }
    }
    for _, f = range dirs {
        k = f.Name()
        if _, ok = data.(map[string]interface{})[k]; ! ok {
            data.(map[string]interface{})[k], err = h.LoadDir(filepath.Join(dir, f.Name()))
            if err != nil {
                return
            }
        }
    }
    return
}

func (h *SyntaxHandlerBD) Write(wr io.Writer, data interface{}) (err error) {
    return nil
}

func (h *SyntaxHandlerBD) Dump(b []byte, data interface{}) (n int, err error) {
    return 0, nil
}
