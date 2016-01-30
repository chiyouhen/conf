// Copyright 2015, 2016 ZHANG Heng (chiyouhen@gmail.com)
//
//  This file is part of Conf.
//  
//  Conf is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  any later version.
//  
//  Conf is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//  
//  You should have received a copy of the GNU General Public License
//  along with Conf.  If not, see <http://www.gnu.org/licenses/>.

package conf

import (
    "os"
    "io"
    "fmt"
    "bytes"
    "bufio"
    "strings"
    "regexp"
)

type SyntaxHandlerHdf struct {
}

var (
    hdfCommentString = map[string]bool{
        "#": true,
        ";": true,
    }
    hdfValidKeyPattern, _ = regexp.Compile("^[A-Za-z0-9\\*]+$")
    hdfDumpIndent = "    "
)

func (h *SyntaxHandlerHdf) Read(rd io.Reader) (data interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    var sc = bufio.NewScanner(rd)
    var l string;
    var stack = make([]map[string]interface{}, 0, 512)
    var curr = data.(map[string]interface{})
    var ok bool
    var k string
    var v interface{}
    var fp *os.File
    var i int

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
        if _, ok = hdfCommentString[string(l[0])]; ok {
            continue
        }

        if strings.HasSuffix(l, "{") {
            // new block
            k = strings.TrimSpace(strings.TrimSuffix(l, "{"))
            if ! hdfValidKeyPattern.Match([]byte(k)) {
                err = &ConfError{"invalid key", filename, lineno}
                return
            }
            v = make(map[string]interface{})
            // auto detect array
            if _, ok = curr[k]; ok {
                switch curr[k].(type) {
                    case []interface{}:
                        curr[k] = append(curr[k].([]interface{}), v)
                    default:
                        curr[k] = []interface{}{curr[k], v}
                }
            } else {
                curr[k] = v
            }
            stack = append(stack, v.(map[string]interface{}))
            curr = v.(map[string]interface{})
        } else if l == "}" {
            // close block
            stack = stack[:len(stack) - 1]
            curr = stack[len(stack) - 1]
        } else {
            // key value pair
            i = strings.Index(l, "=")
            if i < 0 {
                err = &ConfError{"syntax error", filename, lineno}
                return
            }
            k = strings.TrimSpace(string(l[:i]))
            v = strings.TrimSpace(string(l[i + 1:]))
            if ! hdfValidKeyPattern.Match([]byte(k)) {
                err = &ConfError{"invalid key", filename, lineno}
                return
            }
            // auto detect array
            if _, ok = curr[k]; ok {
                switch curr[k].(type) {
                    case []interface{}:
                        curr[k] = append(curr[k].([]interface{}), v)
                    default:
                        curr[k] = []interface{}{curr[k], v}
                }
            } else {
                curr[k] = v
            }
        }
    }
    if len(stack) > 1 {
        err = &ConfError{"syntax error, expect } to close block", filename, lineno}
    }
    return
}

func (h *SyntaxHandlerHdf) Parse(b []byte) (data interface{}, err error) {
    var buf = bytes.NewBuffer(b)
    return h.Read(buf)
}

func (h *SyntaxHandlerHdf) Load(filename string) (data interface{}, err error) {
    var rd *os.File
    rd, err = os.OpenFile(filename, os.O_RDONLY, 0644)
    if err != nil {
        return
    }
    defer rd.Close()
    data, err = h.Read(rd)
    return
}

func (h *SyntaxHandlerHdf) LoadDir(dir string) (data interface{}, err error) {
    err = nil
    data = make(map[string]interface{})
    return
}

func (h *SyntaxHandlerHdf) Write(wr io.Writer, data interface{}) (n int, err error) {
    var b []byte
    b, err = h.Dump(data)
    if err != nil {
        return
    }
    n = 0
    var l = len(b)
    var i = 0
    for n < l {
        i, err = wr.Write(b[n:])
        n += i
        if err != nil && err != io.ErrShortWrite {
            return
        }
    }
    return
}

func (h *SyntaxHandlerHdf) dumpkv(key string, value interface{}, level int) (b []byte, err error) {
    b = make([]byte, 0, 0)
    var bb []byte
    switch value.(type) {
        case map[string]interface{}:
            var k string
            var v interface{}
            bb = []byte(fmt.Sprintf("%s%s {\n", strings.Repeat(hdfDumpIndent, level), key))
            b = append(b, bb...)
            for k, v = range value.(map[string]interface{}) {
                bb, err = h.dumpkv(k, v, level + 1)
                if err != nil {
                    return
                }
                b = append(b, bb...)
            }
            bb = []byte(fmt.Sprintf("%s}\n", strings.Repeat(hdfDumpIndent, level)))
            b = append(b, bb...)
        case []interface{}:
            var v interface{}
            for _, v = range value.([]interface{}) {
                bb, err = h.dumpkv(key, v, level)
                if err != nil {
                    return
                }
                b = append(b, bb...)
            }
        default:
            bb = []byte(fmt.Sprintf("%s%s = %s\n", strings.Repeat(hdfDumpIndent, level), key, value.(string)))
            b = append(b, bb...)
    }
    return
}

func (h *SyntaxHandlerHdf) Dump(data interface{}) (b []byte, err error) {
    b = make([]byte, 0, 0)
    var bb []byte
    switch data.(type) {
        case map[string]interface{}:
            var k string
            var v interface{}
            for k, v = range data.(map[string]interface{}) {
                bb, err = h.dumpkv(k, v, 0)
                if err != nil {
                    return
                }
                b = append(b, bb...)
            }
        default:
            err = &ConfError{"invalid data structure", "", 0}
    }
    return
}
