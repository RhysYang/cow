package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"runtime"
)

// ReadLine read till '\n' is found or encounter error. The returned line does
// not include ending '\r' and '\n'. If returns err != nil if and only if
// len(line) == 0.
func ReadLine(r *bufio.Reader) (line string, err error) {
	line, err = r.ReadString('\n')
	n := len(line)
	if n > 0 && (err == nil || err == io.EOF) {
		id := n - 1
		if line[id] == '\n' {
			id--
		}
		for ; id >= 0 && line[id] == '\r'; id-- {
		}
		return line[:id+1], nil
	}
	return
}

func IsDigit(b byte) bool {
	return '0' <= b && b <= '9'
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func isFileExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.Mode()&os.ModeType == 0 {
			return true, nil
		}
		return false, errors.New(path + " exists but is not regular file")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func isDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, errors.New(path + " exists but is not directory")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Get host IP address
func hostIP() (addrs []string, err error) {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("Error get host name: %v\n", err)
		return
	}

	addrs, err = net.LookupHost(name)
	if err != nil {
		fmt.Printf("Error getting host IP address: %v\n", err)
		return
	}
	return
}

func trimLastDot(s string) string {
	if len(s) > 0 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}

func getUserHomeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		fmt.Println("HOME environment variable is empty")
	}
	return home
}

func expandTilde(pth string) string {
	if len(pth) > 0 && pth[0] == '~' {
		home := getUserHomeDir()
		return path.Join(home, pth[1:])
	}
	return pth
}

func copyN(r io.Reader, w, contBuf io.Writer, n int, buf, pre, end []byte) (err error) {
	var nn int
	bufLen := len(buf)
	var b []byte
	for n != 0 {
		if pre != nil {
			if len(pre) >= bufLen {
				if _, err = w.Write(pre); err != nil {
					return
				}
				pre = nil
				continue
			}
			// append pre to buf
			copy(buf, pre)
			if len(pre)+n < bufLen {
				b = buf[len(pre) : len(pre)+n]
			} else {
				b = buf[len(pre):]
			}
		} else {
			if n < bufLen {
				b = buf[:n]
			} else {
				b = buf
			}
		}
		if nn, err = r.Read(b); err != nil {
			return
		}
		n -= nn
		if pre != nil {
			nn += len(pre)
			pre = nil
		}
		if n == 0 && end != nil && nn+len(end) <= bufLen {
			copy(buf[nn:], end)
			nn += len(end)
			end = nil
		}
		if contBuf != nil {
			contBuf.Write(buf[:nn])
		}
		if _, err = w.Write(buf[:nn]); err != nil {
			return
		}
	}
	if end != nil {
		if _, err = w.Write(end); err != nil {
			return
		}
	}
	return
}

func md5sum(ss ...string) string {
	h := md5.New()
	for _, s := range ss {
		io.WriteString(h, s)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NetNbitIPv4Mask returns a IPMask with highest n bit set.
func NewNbitIPv4Mask(n int) net.IPMask {
	if n > 32 {
		panic("NewNbitIPv4Mask: bit number > 32")
	}
	mask := []byte{0, 0, 0, 0}
	for id := 0; id < 4; id++ {
		if n >= 8 {
			mask[id] = 0xff
		} else {
			mask[id] = ^byte(1<<(uint8(8-n)) - 1)
			break
		}
		n -= 8
	}
	return net.IPMask(mask)
}
