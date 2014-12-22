package P2P

/*
#cgo LDFLAGS: ./lib/AVAPIs.dll ./lib/IOTCAPIs.dll
#cgo CFLAGS: -I./Include

#include <stdio.h>
#include <string.h>
#include <time.h>
#include <stdlib.h>
#include <Winsock2.h>
#include <ws2tcpip.h>
#include <windows.h>
#include <sys/timeb.h>
#include <wchar.h>
#include <io.h>
#pragma comment(lib, "ws2_32")
#define WSA_VERSION MAKEWORD(2, 2) // using winsock 2.2
#include "IOTCAPIs.h"
#include "AVAPIs.h"
#include "AVFRAMEINFO.h"
#include "AVIOCTRLDEFs.h"

#define SERVTYPE_STREAM_SERVER	1
#define AUDIO_STREAM_OUT_CH	1
#define MAX_SIZE_IOCTRL_BUF		1024

#define VIDEO_RECORD_FRAMES 120
#define AUDIO_RECORD_FRAMES 250

#define AUDIO_BUF_SIZE	1024

#define ENABLE_AUDIO 0
#define ENABLE_SPEAKER 0

extern void CArrayToGoArrayG();
extern int addPubchan();
int CIOTC_Initialize(){
	int ret = IOTC_Initialize(0, "61.188.37.216", "50.19.254.134", "m2.iotcplatform.com", "m4.iotcplatform.com");
	 return ret;
}

int CavClientStart2(int SID,char *pw){
	int nResend=-1;
	unsigned long srvType;
	int avIndex = avClientStart2(SID, "admin", pw, 5, &srvType, 0, &nResend);
	return avIndex;
}

int start_ipcam_stream(int avIndex,int* num)
{
	int ret;
	unsigned short val = 0;
	if((ret = avSendIOCtrl(avIndex, IOTYPE_INNER_SND_DATA_DELAY, (char *)&val, sizeof(unsigned short)) < 0))
	{
		printf("start_ipcam_stream failed[%d]\n", ret);
		return 0;
	}
	printf("send Cmd: IOTYPE_INNER_SND_DATA_DELAY, OK\n");

	SMsgAVIoctrlAVStream ioMsg;
	memset(&ioMsg, 0, sizeof(SMsgAVIoctrlAVStream));
	if((ret = avSendIOCtrl(avIndex, IOTYPE_USER_IPCAM_START, (char *)&ioMsg, sizeof(SMsgAVIoctrlAVStream))) < 0)
	{
		printf("start_ipcam_stream failed[%d]\n", ret);
		return 0;
	}
	printf("send Cmd: IOTYPE_USER_IPCAM_START, OK\n");

#if ENABLE_AUDIO
	if((ret = avSendIOCtrl(avIndex, IOTYPE_USER_IPCAM_AUDIOSTART, (char *)&ioMsg, sizeof(SMsgAVIoctrlAVStream))) < 0)
	{
		printf("start_ipcam_stream failed[%d]\n", ret);
		return 0;
	}
	printf("send Cmd: IOTYPE_USER_IPCAM_AUDIOSTART, OK\n");
#endif

#if ENABLE_SPEAKER
	ioMsg.channel = AUDIO_SPEAKER_CHANNEL;
	if((ret = avSendIOCtrl(avIndex, IOTYPE_USER_IPCAM_SPEAKERSTART, (char *)&ioMsg, sizeof(SMsgAVIoctrlAVStream))) < 0)
	{
		printf("start_ipcam_stream failed[%d]\n", ret);
		return 0;
	}
	printf("send Cmd: IOTYPE_USER_IPCAM_SPEAKERSTART, OK\n");
#endif
	*num=addPubchan();
	return 1;
}

#define VIDEO_BUF_SIZE	48000

void thread_ReceiveVideo(int arg,int num)
{
	printf("[thread_ReceiveVideo] Starting....\n");

	int avIndex = arg;
	char buf[VIDEO_BUF_SIZE];
	int ret;

	FRAMEINFO_t frameInfo;
	unsigned int frmNo;
	printf("Start IPCAM video stream OK![%d]\n", avIndex);
	int flag = 0, cnt = 0;
	//char fn[32];
	while(1)
	{
		ret = avRecvFrameData(avIndex, buf, VIDEO_BUF_SIZE, (char *)&frameInfo, sizeof(FRAMEINFO_t), &frmNo);

		if(ret == AV_ER_DATA_NOREADY)
		{
			//printf("AV_ER_DATA_NOREADY[%d]\n", frmNo);
			Sleep(300);
			continue;
		}
		else if(ret == AV_ER_LOSED_THIS_FRAME)
		{
			printf("Lost video frame NO[%d]\n", frmNo);
			continue;
		}
		else if(ret == AV_ER_INCOMPLETE_FRAME)
		{
			printf("Incomplete video frame NO[%d]\n", frmNo);
			continue;
		}
		else if(ret == AV_ER_SESSION_CLOSE_BY_REMOTE)
		{
			printf("[thread_ReceiveVideo] AV_ER_SESSION_CLOSE_BY_REMOTE\n");
			break;
		}
		else if(ret == AV_ER_REMOTE_TIMEOUT_DISCONNECT)
		{
			printf("[thread_ReceiveVideo] AV_ER_REMOTE_TIMEOUT_DISCONNECT\n");
			break;
		}
		else if(ret == IOTC_ER_INVALID_SID)
		{
			printf("[thread_ReceiveVideo] Session cant be used anymore\n");
			break;
		}else{
			CArrayToGoArrayG(buf,ret,frameInfo.flags,num);
		}

	}

	printf("[thread_ReceiveVideo] thread exit\n");

	//return 0;
}


*/
import "C"

import (
	"fmt"
	"github.com/astaxie/beego"
	//	"unsafe"
)

type P2PCON struct {
	uid      string
	pw       string
	avIndex  int
	SID      int
	linkFlag bool
}

const (
	nRTMPType C.uint = 98765
)

func NewP2PCON(uidsrc string, pw string) *P2PCON {
	return &P2PCON{uid: uidsrc, pw: pw, linkFlag: false}
}

func (p *P2PCON) Dial() error {
	fmt.Println("Dialing...")
	count := 0
	var (
		sid, renum int
		err        error
	)
	for count < 3 {
		sid, renum, err = dial(p.uid, p.pw)
		if err == nil {
			break
		}
		count++
	}
	//sid, renum, err := dial(p.uid, p.pw)
	if err != nil {
		return err
	}
	p.SID = sid
	p.avIndex = renum
	p.linkFlag = true
	return nil
}

func (p *P2PCON) Write(iotype C.uint, msg string) error {
	err := rdtwrite(p.avIndex, iotype, msg)
	if err != nil {
		C.IOTC_Session_Close(C.int(p.SID))
		p.linkFlag = false
		return err
	}
	return nil
}

func (p *P2PCON) Read() (string, C.uint, error) {
	str, iotype, err := rdtread(p.avIndex)
	if err != nil {
		C.IOTC_Session_Close(C.int(p.SID))
		p.linkFlag = false
		return "", C.uint(0), err
	}
	return str, iotype, nil
}

func (p *P2PCON) BroadCast() error {
	//ch := make(chan *Message, 1)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			//log.Info("%v", err)
		}
	}()
	var num C.int
	ret := C.start_ipcam_stream(C.int(p.avIndex), &num)
	if int(ret) < 0 {
		return fmt.Errorf("start_ipcam_stream error")
	}
	//m := &iMessage{byte(0x09), []byte{0x17, 0, 0, 0, 0, 0x1, 0x4D, 0x00, 0x1E, 0xFF, 0xE1, 0, 0x0A, 0x67, 0x4D, 0x00, 0x1E, 0x95, 0xA8, 0x28, 0x0B, 0xFE, 0x54, 0x1, 0, 0x04, 0x68, 0xEE, 0x3C, 0x80}}
	//videochan <- m
	ch := pubchanslice[int(num)]
	go broadCast(ch, "rtmp://192.168.8.135/live", "eids2")
	C.thread_ReceiveVideo(C.int(p.avIndex), num)
	return nil
}

func (p *P2PCON) Close() {
	myclose(p.avIndex, p.SID)
	p.linkFlag = false
}
func init() {
	//beego.SetLogger("file", `{"filename":"test.log"}`)
	//beego.SetLevel(3)
	C.IOTC_DeInitialize()
	//log := logs.NewLogger(10000)
	//log.SetLogger("file", `{"filename":"test.log"}`)

	ret := C.CIOTC_Initialize()
	//beego.Info(ret)
	if int(ret) < 0 {
		//log.Error("IOTC_Initialize error!!")
		beego.Info("IOTC_Initialize error!!")
	}
	C.avInitialize(C.int(32))
	//log.Info("init finish")

}

func Uinit() {
	C.avDeInitialize()
	C.IOTC_DeInitialize()
}

func dial(uid string, pw string) (int, int, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			//log.Info("%v", err)
		}
	}()
	//beego.Info(uid)
	tmpSID := C.IOTC_Get_SessionID()
	if int(tmpSID) < 0 {
		return 0, 0, fmt.Errorf("IOTC_Get_SessionID failed[%d]\n", int(tmpSID))
	}
	SID := C.IOTC_Connect_ByUID_Parallel(C.CString(uid), tmpSID)
	beego.Info("Step 1: call IOTC_Connect_ByUID(", uid, ") ret(", int(SID), ").......")
	//beego.Info(uid+">>SID=", SID)
	if int(SID) < 0 {

		//log.Info("p2pAPIs_Client connect failed...!!")
		return 0, 0, fmt.Errorf("p2p_connect failed!")
	}

	avIndex := C.CavClientStart2(SID, C.CString(pw))
	reNum := int(avIndex)
	beego.Info("Step 2: call avClientStart2(", int(SID), ") ret(", reNum, ")......")
	if reNum < 0 {
		//log.Info("avClientStart failed!!", reNum)
		C.IOTC_Session_Close(SID)
		return 0, 0, fmt.Errorf("avClientStart failed!")
	}
	beego.Info("avClientStart2 OK")
	return int(SID), reNum, nil

}

func rdtwrite(avIndex int, ioType C.uint, msg string) error {
	cmsg := C.CString(msg)
	lenm := len(msg)
	ret := C.avSendIOCtrl(C.int(avIndex), ioType, cmsg, C.int(lenm))
	if ret < 0 {
		return fmt.Errorf("avSendIOCtrl send cmd failed[%d]!!", int(ret))
	}
	return nil
}

func rdtread(avIndex int) (string, C.uint, error) {
	var buf *C.char
	var reType C.uint
	ret := C.avRecvIOCtrl(C.int(avIndex), &reType, buf, 1024, C.uint(1000))
	if ret < 0 {
		return "", 0, fmt.Errorf("avSendIOCtrl recive cmd failed[%d]!!", int(ret))
	}
	return C.GoString(buf), reType, nil
}

func myclose(avIndex int, SID int) {
	C.avClientStop(C.int(avIndex))
	C.IOTC_Session_Close(C.int(SID))
}
