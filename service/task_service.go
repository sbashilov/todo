package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sbashilov/todo/pb"
	"google.golang.org/protobuf/proto"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

const (
	sizeLen = 8
	file    = "tasks.pb"
)

// TaskService grpc server impl
type TaskService struct {
	pb.UnimplementedTaskServiceServer
}

// Add add new todo task to list
func (ts *TaskService) Add(ctx context.Context, t *pb.Task) (*empty.Empty, error) {
	if t == nil {
		return nil, errors.New("string required")
	}
	b, err := proto.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("could not marshal message: %v", err)
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("could not open a file: %s, %v", file, err)
	}

	if err := binary.Write(f, binary.LittleEndian, int64(len(b))); err != nil {
		return nil, fmt.Errorf("could not write len to file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		return nil, fmt.Errorf("could not write data to file: %v", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("could not close file: %v", err)
	}
	return &empty.Empty{}, nil
}

// List returns list of tasks
func (ts *TaskService) List(ctx context.Context, _ *empty.Empty) (*pb.Tasks, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while reading file: %v", err)
	}
	tasks := &pb.Tasks{
		Tasks: make([]*pb.Task, 0),
	}
	for {
		if len(b) == 0 {
			break
		} else if len(b) < sizeLen {
			return nil, fmt.Errorf("remaining odd %d bytes", len(b))
		}
		var l int64
		if err := binary.Read(bytes.NewReader(b[:sizeLen]), binary.LittleEndian, &l); err != nil {
			return nil, fmt.Errorf("could not decode message length: %v", err)
		}
		b = b[sizeLen:]
		task := &pb.Task{}
		if err := proto.Unmarshal(b[:l], task); err != nil {
			return nil, fmt.Errorf("could not decode message: %v", err)
		}
		b = b[l:]
		tasks.Tasks = append(tasks.Tasks, task)
	}
	return tasks, nil
}

// ListPrint prints tasks list to stdout
func (ts *TaskService) ListPrint() error {
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
}

// Done marks task done by number
func (ts *TaskService) Done(tNum string) error {
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
			task := &pb.Task{}
			if err := proto.Unmarshal(b[:l], task); err != nil {
				return fmt.Errorf("could not decode message: %v", err)
			}
			task.Done = true
			return ts.updateMsg(task, bytesPassed, b[l:])
		}
		bytesPassed += sizeLen
		b = b[l:]
		bytesPassed += l
	}
}

func (ts *TaskService) updateMsg(t *pb.Task, from int64, rest []byte) error {
	b, err := proto.Marshal(t)
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
