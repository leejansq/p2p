package P2P

import (
	//"flag"
	"fmt"
	//"github.com/zhangpeihao/flv"
	"github.com/zhangpeihao/log"
	"github.com/zhangpeihao/rtmp"
	"os"
	"time"
)

const (
	programName = "RtmpPublisher"
	version     = "0.0.1"
)

type STestOutboundConnHandler struct {
	ch     chan *iMessage
	url    string
	stream string
	obConn rtmp.OutboundConn
	status uint
}

var l *log.Logger = log.NewLogger(".", "publisher", nil, 60, 3600*24, true)

//var obConn rtmp.OutboundConn
//var status uint

func (handler *STestOutboundConnHandler) OnStatus(conn rtmp.OutboundConn) {
	//var err error
	handler.status, _ = handler.obConn.Status()
	//fmt.Printf("@@@@@@@@@@@@@status: %d, err: %v\n", status, err)
}

func (handler *STestOutboundConnHandler) OnClosed(conn rtmp.Conn) {
	fmt.Printf("@@@@@@@@@@@@@Closed\n")

}

func (handler *STestOutboundConnHandler) OnReceived(conn rtmp.Conn, message *rtmp.Message) {
}

func (handler *STestOutboundConnHandler) OnReceivedRtmpCommand(conn rtmp.Conn, command *rtmp.Command) {
	fmt.Printf("ReceviedRtmpCommand: %+v\n", command)
}

func (handler *STestOutboundConnHandler) OnStreamCreated(conn rtmp.OutboundConn, stream rtmp.OutboundStream) {
	fmt.Printf("Stream created: %d\n", stream.ID())
	stream.Attach(handler)
	err := stream.Publish(handler.stream, "live")
	if err != nil {
		fmt.Printf("Publish error: %s", err.Error())
		os.Exit(-1)
	}
}
func (handler *STestOutboundConnHandler) OnPlayStart(stream rtmp.OutboundStream) {

}
func (handler *STestOutboundConnHandler) OnPublishStart(stream rtmp.OutboundStream) {
	// Set chunk buffer size
	fmt.Printf("Stream publishStart: %d\n", stream.ID())
	go handler.publish_lee(stream)
}

func (handler *STestOutboundConnHandler) publish_lee(stream rtmp.OutboundStream) {
	defer func() {
		//l.Close()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		handler.obConn.Close()
	}()
	ch := handler.ch
	startTs := int64(0)
	diff := uint32(0)
	for handler.status == rtmp.OUTBOUND_CONN_STATUS_CREATE_STREAM_OK {

		select {
		case message := <-ch:
			fmt.Println(message.tagtype)
			if startTs == 0 {
				startTs = time.Now().UnixNano()
			} else {
				diff = uint32((time.Now().UnixNano() - startTs) / 1000000)
			}
			//fmt.Println(message.tagtype, message.data[:10], diff)
			if err := stream.PublishData(message.tagtype, message.data, diff); err != nil {
				fmt.Println("PublishData() error:", err)
				break
			}

		case <-time.After(10 * time.Second):
			break
		}

	}

}

func broadCast(ch chan *iMessage, outch chan string, url string, stream string) {
	//l = log.NewLogger(".", "publisher", nil, 60, 3600*24, true)
	rtmp.InitLogger(l)
	defer func() {
		//l.Close()
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	outHandler := &STestOutboundConnHandler{ch, url, stream, nil, 0}
	fmt.Println("to dial")
	var err error
	obConn, err := rtmp.Dial(url, outHandler, 100)
	outHandler.obConn = obConn
	if err != nil {
		fmt.Println("Dial error", err)
		os.Exit(-1)
	}
	//defer obConn.Close()
	fmt.Println("to connect")
	err = obConn.Connect()
	if err != nil {
		fmt.Printf("Connect error: %s", err.Error())
		os.Exit(-1)
	}
	fmt.Println("broadcast init finished!")
	tmpstr := url + "/" + stream
	outch <- tmpstr
	//time.Sleep(time.Hour)
}
