package P2P

import (
	"babylon/rtmp"
	"fmt"
	//"net"
	//"os"
	"runtime"
	//"strings"
	//"bufio"
	//"os"
	"testing"
)

func Test_A(t *testing.T) {

	runtime.GOMAXPROCS(runtime.NumCPU())
	//conn, err := net.Dial("udp", "baidu.com:80")
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return
	//}
	//fmt.Println(conn.LocalAddr())
	//Localaddr = strings.Split(conn.LocalAddr().String(), ":")[0]
	//conn.Close()
	go func() {
		err := rtmp.ListenAndServe(":1935")
		if err != nil {
			panic(err)
		}
	}()
	//ch := make(chan int, 1)ERNTVKNP4WU7AM3USR5J,jDY3FIRsmqL2MVN
	//go broadCast(videochan, "rtmp://192.168.8.135/live", "eids")
	fmt.Println("ASD")
	//go func() {
	obj := NewP2PCON("ERNTVKNP4WU7AM3USR5J", "888888")
	err := obj.Dial()
	if err != nil {
		fmt.Println(err)
		//ch <- 1
		return
	}
	fmt.Println(obj)
	obj.BroadCast()
	//startReceiveVideo(obj.avIndex)
	//obj.Write(nRTMPType, "AAA")
	//obj.Close()
	//	ch <- 1
	//}()
	//obj := NewP2PCON("DRHTVKME9WUF9G3US7C1", "888888")
	//err := obj.Dial()
	//if err != nil {
	//	fmt.Println("ASD", <-ch)
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(obj)
	////obj.Write(nRTMPType, "AAA")

	//fmt.Println("ASD", <-ch)
	//obj.Close()

	//file, err := os.Open("wait.g726")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//rdr := bufio.NewReader(file)
	//b := make([]byte, 80)
	//for {
	//	var n int
	//	n, err = rdr.Read(b)
	//	fmt.Println(n)

	//	//fmt.Println(n)

	//	if err != nil {
	//		fmt.Println(err)
	//		break
	//		//	}
	//	} else {
	//		obj.Speak(b[:n])
	//	}
	//}

	obj.Write(uint32(0x1300), "atsd")
	for {
		str, its, _ := obj.Read()
		fmt.Println(its)
		if uint32(its) == uint32(0x1301) {
			fmt.Println("<<<<<<<<<<", str)
			break
		}

	}
	//str, _, _ := obj.Read()
	//fmt.Println("<<<<<<<<<<", str)
}
