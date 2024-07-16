package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	client := http.Client{}
	client.Transport = HandleAeuthnticate(HandleLogger(http.DefaultTransport), os.Getenv("APP_ID"), os.Getenv("TOKEN"))
	_, err := client.Get("https://sandbox.api.sgroup.qq.com/users/@me")
	if err != nil {
		panic(err)
	}

}

type TransportFunc func(req *http.Request) (*http.Response, error)

func (f TransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func HandleLogger(transport http.RoundTripper) http.RoundTripper {
	return TransportFunc(func(req *http.Request) (*http.Response, error) {
		r, err := httputil.DumpRequest(req, true)
		if err != nil {
			return nil, fmt.Errorf("dump request failed: %w", err)
		}
		log.Println(string(r))
		resp, rerr := transport.RoundTrip(req)
		if rerr != nil {
			return resp, rerr
		}
		data, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, fmt.Errorf("dump response failed: %w", err)
		}
		log.Println(string(data))
		return resp, nil
	})
}

func HandleAeuthnticate(transport http.RoundTripper, appID string, token string) http.RoundTripper {
	signal := struct {
		AccessToken string
		Expire      time.Time
	}{}
	lock := sync.RWMutex{}
	return TransportFunc(func(req *http.Request) (*http.Response, error) {
		lock.RLock()
		if signal.Expire.Before(time.Now()) {
			lock.RUnlock()
			lock.Lock()
			client := http.Client{
				Transport: transport,
			}
			resp, err := client.Post("https://bots.qq.com/app/getAppAccessToken",
				"application/json",
				strings.NewReader(`{"appId":"`+appID+`","clientSecret":"`+token+`"}`))
			if err != nil {
				return nil, fmt.Errorf("get access token failed: %w", err)
			}
			structureBody := struct {
				AccessToken string `json:"access_token"`
				ExpiresIn   string `json:"expires_in"`
			}{}
			defer resp.Body.Close()
			if err := json.NewDecoder(resp.Body).Decode(&structureBody); err != nil {
				return nil, fmt.Errorf("decode response failed: %w", err)
			}
			expiresIn, err := strconv.ParseInt(structureBody.ExpiresIn, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse expires_in failed: %w", err)
			}
			signal.AccessToken = structureBody.AccessToken
			signal.Expire = time.Now().Add(time.Second * time.Duration(expiresIn))
			lock.Unlock()
			lock.RLock()
		}
		req.Header.Set("Authorization", "QQBot "+signal.AccessToken)
		req.Header.Set("X-Union-Appid", appID)
		lock.RUnlock()
		return transport.RoundTrip(req)
	})
}
