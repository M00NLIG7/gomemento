package common

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"os"
	"strings"
	"os/exec"

	"github.com/klauspost/compress/zstd"
)

type finfo struct {
	Name string
	Size int64
	Time string
	Hash string
}

func CheckFile(name string) finfo {
	fileInfo, err := os.Stat(name)
	if err != nil {
		panic(err)
	}
	println(name)
	if fileInfo.IsDir() {

		t := fileInfo.ModTime().String()
		b := fileInfo.Size()

		i := finfo{
			Name: name,
			Size: b,
			Time: t,
			Hash: "directory",
		}

		return i
	} else {
		f, err := os.Open(name)
		if err != nil {
			panic(err)
		}
		if err != nil {
			if os.IsNotExist(err) {
				println("file not found:", fileInfo.Name())
			}
		}
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			panic(err)
		}
		hash := h.Sum(nil)
		Enc := base64.StdEncoding.EncodeToString(hash)

		t := fileInfo.ModTime().String()
		b := fileInfo.Size()

		i := finfo{
			Name: name,
			Size: b,
			Time: t,
			Hash: Enc,
		}
		return i
	}
}

func Compress(in io.Reader, out io.Writer) error {
	enc, err := zstd.NewWriter(out)
	if err != nil {
		return err
	}
	//gets data from in and writes it to enc, which is out
	_, err = io.Copy(enc, in)
	if err != nil {
		enc.Close()
		return err
	}
	return enc.Close()
}

func Decompress(in io.Reader, out io.Writer) error {
	d, err := zstd.NewReader(in)
	if err != nil {
		return err
	}
	defer d.Close()

	// Copy content...
	_, err = io.Copy(out, d)
	return err
}

func GetHistFile(username string, shellname string, homedir string) string {
	// for the future, refernce the $HISTFILE variable for each users env
	switch {
	case strings.Contains(shellname, "bash") || strings.Contains(shellname, "sh"):
		shellpathfull := homedir + "/.bash_history"
		return shellpathfull
	case strings.Contains(shellname, "ash"):
		shellpathfull := homedir + "/.ash_history"
		return shellpathfull
	case strings.Contains(shellname, "zsh"):
		shellpathfull := homedir + "/.zsh_history"
		return shellpathfull
	case strings.Contains(shellname, "fish"):
		shellpathfull := homedir + "/.local/share/fish/fish_history"
		return shellpathfull
	}
	return "shell not found"
}

func GetShell() string {
	// TODO: Implement error handling for exec.Command
	shellBytes, _ := exec.Command("bash", "-c", "echo $0").Output()
	return string(shellBytes)
}

func GetHomeDir() string {
	homeDirBytes, _ := exec.Command("bash", "-c", "echo $HOME").Output()
	return string(homeDirBytes)

}

func OpenFile(file string) []string {
	var s []string
	stats := CheckFile(file)
	if stats.Size != 0 {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		// remember to close the file at the end of the program
		defer f.Close()

		// read the file line by line using scanner
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			// do something with a line
			s = append(s, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		//print slice with contents of file
		//for _, str := range s {
		//	println(str)
		//}
	}
	return s
}

func IsHumanReadable(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b := make([]byte, 4)
	_, err = f.Read(b)
	if err != nil {
		panic(err)
	}
	// text := string(b)
	// 60 = E, 76 = L, 70 = F
	// ELF check
	if (b[1] == 69) && (b[2] == 76) && (b[3] == 70) {
		return false
	}
	// check for crazy ass bytes in file
	// 0x00 = null byte
	// 0x01 = start of heading
	// 0x02 = start of text

	r := bufio.NewReader(f)
	for {
		if c, _, err := r.ReadRune(); err != nil {
			if err == io.EOF {
				break
			}
		} else {
			if string(c) == "\x00" || string(c) == "\x01" || string(c) == "\x02" {
				return false
			}
		}
	}
	return true
}
