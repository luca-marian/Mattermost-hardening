package main

import (
    "time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random_ms_duration(max int) time.Duration {	
	return time.Duration(rand.Intn(max)) * time.Millisecond
}

func retry(attempts int, delay time.Duration, fn func() error) error {
    var err error
    for i := 0; i < attempts; i++ {
        if err = fn(); err == nil {
            return nil
        }
        time.Sleep(delay)
    }
    return err
}