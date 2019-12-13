package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Type = flag.String("type", "server", "input run type, client or server")
var Port = flag.String("port", "8500~8503", "input port range , 8500~8503 meaning port from 8500 to 8530")
var Size = flag.Int("size", 1024, "package size")
var Time = flag.Int("time", 60, "test time")
var Addr = flag.String("address", "127.0.0.1", "server addr")
var Count = flag.Int("count", 100, "client count")

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func MD5(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

func RandStringBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}

func Server(addr string, port uint16, wg *sync.WaitGroup) {
	defer wg.Done()
	if addr == "127.0.0.1" || addr == "0.0.0.0" {
		addr = ""
	}
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, port))
	log.Print(err)

	packageCount  := 0
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		fmt.Print(err)
	}
	defer conn.Close()
	for {

		time.Sleep(300 * time.Millisecond)
		buf := make([]byte , 2048)
		length, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil  {
			log.Printf("read from %s , lenght %d", remoteAddr.String(), length)
		}

		if length == 0 {
			log.Printf("reciver bytes : %d " , length)
			continue
		}
		packageCount++
		if packageCount%100 == 0{
			log.Printf("package count: %d", packageCount)
		}
		if err != nil {
			length, err := conn.WriteToUDP(buf, remoteAddr)
			if err != nil {
				log.Printf("write to %s , lenght %d", remoteAddr.String(), length)
			} else {
				log.Printf("write to %s , lenght %d  faile", remoteAddr.String(), length)
			}
		}
	}
}

func Client(addr string, port uint16, index int, ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.Dial("udp", fmt.Sprintf("%s:%d", addr, port))
	defer conn.Close()
	if err != nil {
		fmt.Printf("client %d dial fiale", index)
		return
	}
	packageCount  := 0
	for {
		var randomString = RandStringBytes(1024)
		md := MD5(string(randomString))
		body := bytes.Join([][]byte{Int64ToBytes(int64(index)), []byte(md), randomString}, []byte(""))
		lenght, err := conn.Write(body)
		if err != nil {
			log.Printf("write %d dial fiale", index)
		} else {
			log.Printf("write %d wite to %s:%d suss , len %d",index, addr , port ,lenght)
		}

		var buf []byte
		lenght, err = conn.Read(buf)
		packageCount++
		if packageCount%100 == 0{
			log.Printf("package count: %d", packageCount)
		}
		if err != nil {
			log.Printf("read %d dial fiale", index)
		} else {
			log.Printf("read %d dial suss , len %d", index, lenght)
		}

		select {
		case <-ctx.Done():
			log.Printf("client %d out", index)
			return
		default:
			log.Printf("client %d on", index)
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func main() {
	flag.Parse()
	log.Printf("type: %s , port: %s , size: %d , time: %d", *Type, *Port, *Size, *Time)

	ports := strings.Split(*Port, "~")
	portStart, _ := strconv.ParseUint(ports[0], 10, 16)
	portEnd, _ := strconv.ParseUint(ports[1], 10, 16)
	wg := sync.WaitGroup{}
	if *Type == "server" {
		for i := portStart; i < portEnd+1; i++ {
			go Server(*Addr, uint16(i), &wg)
			wg.Add(1)
		}
	} else if *Type == "client" {
		var portArr []uint16
		for i := portStart; i < portEnd+1; i++ {
			portArr = append(portArr, uint16(i))
		}

		for i := 0; i < *Count; i++ {
			port := portArr[i%len(portArr)]
			timeout := *Time
			var ctx, _ = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			go Client(*Addr, port, i, ctx, &wg)
			wg.Add(1)
		}

	} else {
		log.Println("Server Type Error")
	}

	wg.Wait()
}
