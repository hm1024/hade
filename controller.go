package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hm1024/hade/framework"
)

func FooControllerHandler(c *framework.Context) error {
	finish := make(chan struct{}, 1)
	// 这个 channel 负责通知 panic 异常
	panicChan := make(chan interface{}, 1)

	durationCtx, cancel := context.WithTimeout(c.BaseContext(), time.Duration(1*time.Second))
	defer cancel()

	go func() {
		defer func() {
			if p := recover(); p != nil {
				panicChan <- p
			}
		}()

		time.Sleep(10 * time.Second)
		c.WriterMux().Lock()
		defer c.WriterMux().Unlock()
		c.Json(http.StatusOK, "ok")

		finish <- struct{}{}
	}()

	select {
	case p := <-panicChan:
		// ...
		log.Println("panic occurs: ", p)

		c.WriterMux().Lock()
		defer c.WriterMux().Unlock()
		c.Json(http.StatusInternalServerError, "panic")
	case <-finish:
		log.Println("finish")
	case <-durationCtx.Done():
		log.Println("time out")
		c.WriterMux().Lock()
		defer c.WriterMux().Unlock()
		c.Json(http.StatusInternalServerError, "time out")
	}

	return c.Json(200, map[string]interface{}{
		"code": 0,
	})
}
