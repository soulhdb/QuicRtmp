package main

import (
	"net"
	//        "bufio"
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"time"
	//	"strconv"
	quicconn "github.com/marten-seemann/quic-conn"
	rtmp "github.com/zhangpeihao/gortmp"
)

type TestOutboundConnHandler struct {
}

var obConn rtmp.OutboundConn
var createStreamChan chan rtmp.OutboundStream
var videoDataSize int64
var audioDataSize int64

//var flvFile *flv.File
var status uint
var conn net.Conn
var network bytes.Buffer
var enc *gob.Encoder

var (
	url        *string = flag.String("URL", "rtmp://84.20.77.50/live", "The rtmp url to connect.")
	streamName *string = flag.String("Stream", "livestream1", "Stream name to play.")
)

func (handler *TestOutboundConnHandler) OnStatus(conn rtmp.OutboundConn) {
	var err error
	status, err = conn.Status()
	fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", status, err)
}

func (handler *TestOutboundConnHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")
}

type P struct {
	Buf               []byte
	Type              uint8
	Timestamp         uint32
	AbsoluteTimestamp uint32
}

func (handler *TestOutboundConnHandler) OnReceived(rconn rtmp.Conn, message *rtmp.Message) {
	switch message.Type {
	case rtmp.VIDEO_TYPE:
		//		if flvFile != nil {
		//			flvFile.WriteVideoTag(message.Buf.Bytes(), message.AbsoluteTimestamp)
		//		}
		videoDataSize += int64(message.Buf.Len())
		//		fmt.Println("**",message.Buf.Len())
		if err := enc.Encode(P{message.Buf.Bytes(), message.Type, message.Timestamp, message.AbsoluteTimestamp}); err != nil {
			fmt.Println("error")
			panic(err)
		}
	case rtmp.AUDIO_TYPE:
		//		if flvFile != nil {
		//			flvFile.WriteAudioTag(message.Buf.Bytes(), message.AbsoluteTimestamp)
		//		}
		audioDataSize += int64(message.Buf.Len())
		//		fmt.Println("**",message.Buf.Len())
		if err := enc.Encode(P{message.Buf.Bytes(), message.Type, message.Timestamp, message.AbsoluteTimestamp}); err != nil {
			fmt.Println(err)
			panic(err)
		}
	}

}

func (handler *TestOutboundConnHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedCommand: %+v\n", command)
}

func (handler *TestOutboundConnHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	createStreamChan <- stream
}

func main() {
	// utils.SetLogLevel(utils.LogLevelDebug)

	startClient := flag.Bool("c", false, "client")
	flag.Parse()

	if *startClient {
		// run the client
		go func() {
			var err error
			tlsConf := &tls.Config{} //InsecureSkipVerify: true}
			conn, err = quicconn.Dial("streemtechnology.com:8081", tlsConf)
			if err != nil {
				panic(err)
			}
			enc = gob.NewEncoder(conn)
			//        fmt.Fprintf(conn, message+strconv.Itoa(i)+"\n")
			createStreamChan = make(chan rtmp.OutboundStream)
			testHandler := &TestOutboundConnHandler{}
			fmt.Println("to dial")

			obConn, err = rtmp.Dial(*url, testHandler, 100)
			if err != nil {
				fmt.Println("Dial error", err)
				os.Exit(-1)
			}

			defer obConn.Close()
			fmt.Printf("obConn: %+v\n", obConn)
			fmt.Printf("obConn.URL(): %s\n", obConn.URL())
			fmt.Println("to connect")
			err = obConn.Connect()
			if err != nil {
				fmt.Printf("Connect error: %s", err.Error())
				os.Exit(-1)
			}
			for {
				select {
				case stream := <-createStreamChan:
					// Play
					err = stream.Play(*streamName, nil, nil, nil)
					if err != nil {
						fmt.Printf("Play error: %s", err.Error())
						os.Exit(-1)
					}

				case <-time.After(1 * time.Second):
					//				fmt.Printf("Audio size: %d bytes; Vedio size: %d bytes\n", audioDataSize, videoDataSize)
				}
			}
		}()
	}

	time.Sleep(time.Hour)
}
