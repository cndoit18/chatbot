package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/token"
)

func must[T any](u T, err error) T {
	if err != nil {
		panic(err)
	}
	return u
}

func main() {
	token := token.BotToken(must(strconv.ParseUint(os.Getenv("APP_ID"), 10, 32)), os.Getenv("TOKEN"))
	api := botgo.NewOpenAPI(token).WithTimeout(3 * time.Second)
	ctx := context.Background()

	ws, err := api.WS(ctx, nil, "")
	log.Printf("%+v, err:%v", ws, err)

	me, err := api.Me(ctx)
	log.Printf("%+v, err:%v", me, err)
}
