package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/sbashilov/todo/pb"
)

const (
	sizeLen = 8
	file    = "tasks.pb"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "require subcommand")
		os.Exit(1)
	}

	var err error
	switch cmd := flag.Arg(0); cmd {
	case "add":
		err = add(strings.Join(flag.Args()[1:], " "))
	case "list":
		err = list()
	case "done":
		err = done(flag.Arg(1))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func add(text string) error {
	t := &pb.Task{
		Text: text,
	}
	b, err := proto.Marshal(t)
	if err != nil {
		return fmt.Errorf("could not marshal message: %v", err)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("could not open a file: %s, %v", file, err)
	}

	if err := binary.Write(f, binary.LittleEndian, int64(len(b))); err != nil {
		return fmt.Errorf("could not write len to file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("could not write data to file: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close file: %v", err)
	}
	return nil
}

func list() error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error while reading file: %v", err)
	}
	for {
		if len(b) == 0 {
			return nil
		} else if len(b) < sizeLen {
			return fmt.Errorf("remaining odd %d bytes", len(b))
		}
		var l int64
		if err := binary.Read(bytes.NewReader(b[:sizeLen]), binary.LittleEndian, &l); err != nil {
			return fmt.Errorf("could not decode message length: %v", err)
		}
		b = b[sizeLen:]
		var task pb.Task
		if err := proto.Unmarshal(b[:l], &task); err != nil {
			return fmt.Errorf("could not decode message: %v", err)
		}
		b = b[l:]
		if task.Done {
			fmt.Printf("👍")
		} else {
			fmt.Printf("😱")
		}
		fmt.Printf(" %s\n", task.Text)
	}
	return nil
}

func done(tNum string) error {
	n, err := strconv.ParseInt(tNum, 10, 64)
	if err != nil {
		return fmt.Errorf("argument must be int")
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error while reading file: %v", err)
	}
	var bytesPassed int64
	var num int64 = 0
	for {
		num++
		if len(b) == 0 {
			return nil
		} else if len(b) < sizeLen {
			return fmt.Errorf("remaining odd %d bytes", len(b))
		}
		var l int64
		if err := binary.Read(bytes.NewReader(b[:sizeLen]), binary.LittleEndian, &l); err != nil {
			return fmt.Errorf("could not decode message length: %v", err)
		}
		b = b[sizeLen:]
		if n == num {
			var task pb.Task
			if err := proto.Unmarshal(b[:l], &task); err != nil {
				return fmt.Errorf("could not decode message: %v", err)
			}
			task.Done = true
			return updateMsg(task, bytesPassed, b[l:])
		}
		bytesPassed += sizeLen
		b = b[l:]
		bytesPassed += l
	}
}

func updateMsg(t pb.Task, from int64, rest []byte) error {
	b, err := proto.Marshal(&t)
	if err != nil {
		return fmt.Errorf("could not marshal message: %v", err)
	}
	f, err := os.OpenFile(file, os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("could not open a file: %s, %v", file, err)
	}
	if _, err := f.Seek(from, 1); err != nil {
		return fmt.Errorf("could not set pointer to %d, %v", from, err)
	}
	if err := binary.Write(f, binary.LittleEndian, int64(len(b))); err != nil {
		return fmt.Errorf("could not write len to file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("could not write new data to file: %v", err)
	}
	if _, err := f.Write(rest); err != nil {
		return fmt.Errorf("could not write rest data to file: %v", err)
	}
	return nil
}
