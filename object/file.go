package object

import (
	"bufio"
	"fmt"
	"os"
)

type File struct {
	Filename string
	Handle   *os.File
	Reader   *bufio.Reader
	Writer   *bufio.Writer
}

func (o *File) Type() ObjectType { return FILE_OBJ }
func (o *File) Inspect() string {
	return fmt.Sprintf("<file:%s>", o.Filename)
}

func (f *File) Open() error {
	file, err := os.OpenFile(f.Filename, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	f.Handle = file
	f.Reader = bufio.NewReader(file)
	return nil
}

func (f *File) Read() Object {
	if f.Reader == nil {
		return &String{Value: ""}
	}

	s, err := f.Reader.ReadString('\n')
	if err != nil {
		return &String{Value: ""}
	}

	return &String{Value: s}
}

func (f *File) ReadAll() Object {
	if f.Reader == nil {
		return &String{Value: ""}
	}

	var lines []string

	scanner := bufio.NewScanner(f.Handle)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	l := len(lines)
	result := make([]Object, l)

	for i, s := range lines {
		result[i] = &String{Value: s}
	}
	return &Array{Elements: result}

}

func (f *File) Close() Object {
	f.Handle.Close()
	f.Handle = nil
	return &Boolean{Value: true}
}
