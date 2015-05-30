package elog

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

type Counter struct {
	Id    string
	Value int
}

func NewCounter() (interface{}, error) {
	c := &Counter{
		Id:    strconv.FormatInt(time.Now().Unix(), 10),
		Value: 0,
	}

	return c, nil
}

func (c *Counter) Incr(i int) {
	c.Value += i
}

func (c *Counter) GetInfo() {
	fmt.Println("Id:", c.Id, ", Value:", c.Value)
}

func TestSingleton(t *testing.T) {
	var counterSingleton = NewSingleton(NewCounter)

	for i := 0; i < 10; i++ {
		go func() {
			c, err := counterSingleton.Get()
			if c == nil {
				t.Error("c is nil")
			}

			if err != nil {
				t.Error(err.Error())
			}

			c.(*Counter).GetInfo()
			c.(*Counter).Incr(1)
		}()
	}

	runtime.Gosched()
}
