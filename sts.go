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
#define AUDIO_SPEAKER_CHANNEL 4
#define AUDIO_BUF_SIZE	1024

#define ENABLE_AUDIO 1
#define ENABLE_SPEAKER 1

extern void CArrayToGoArrayG();
extern void HotAudio();
extern int addPubchan();
//extern quitBro();
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
	unsigned short val = 1;
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

#define AUDIO_BUF_SIZE	1024
void thread_ReceiveAudio(int arg,int num)
{
	printf("[thread_ReceiveAudio] Starting....\n");
	int flagaf=1;
	int avIndex = arg;
	char buf[AUDIO_BUF_SIZE];

	FRAMEINFO_t frameInfo;
	unsigned int frmNo;
	int recordCnt = 0;
	int ret;
	printf("Start IPCAM audio stream OK![%d]\n", avIndex);

	while(1)
	{

			ret = avCheckAudioBuf(avIndex);
			if(ret < 0) break;
			if(ret < 30) // determined by audio frame rate
			{
				Sleep(100);
				continue;
			}

			ret = avRecvAudioData(avIndex, buf, AUDIO_BUF_SIZE, (char *)&frameInfo, sizeof(FRAMEINFO_t), &frmNo);

			if(ret == AV_ER_SESSION_CLOSE_BY_REMOTE)
			{
				printf("[thread_ReceiveAudio] AV_ER_SESSION_CLOSE_BY_REMOTE\n");
				break;
			}
			else if(ret == AV_ER_REMOTE_TIMEOUT_DISCONNECT)
			{
				printf("[thread_ReceiveAudio] AV_ER_REMOTE_TIMEOUT_DISCONNECT\n");
				break;
			}
			else if(ret == IOTC_ER_INVALID_SID)
			{
				printf("[thread_ReceiveAudio] Session cant be used anymore\n");
				break;
			}
			else if(ret == AV_ER_LOSED_THIS_FRAME)
			{
				printf("AV_ER_LOSED_THIS_FRAME[%d]\n", frmNo);
				continue;
			}
			else if(ret < 0)
			{
				printf("Other error[%d]!!!\n", ret);
				continue;
			}
			else
			{
				HotAudio(buf,ret,flagaf,num);
				flagaf=0;
			}

			//audio_playback(fd, buf, ret);
			//printf("[%d]", ret); //fflush(stdout);
			//if(recordCnt++ > AUDIO_RECORD_FRAMES) break;

	}

	printf("[thread_ReceiveAudio] thread exit\n");

	//return 0;
}
#define VIDEO_BUF_SIZE	300000
int gSleepMs = 30;

void thread_ReceiveVideo(int arg,int num)
{
//PacRec();
printf("[thread_ReceiveVideo] Starting....\n");
	int avIndex = arg;
	char buf[VIDEO_BUF_SIZE];
	int ret=0;
	//int fd = open_videoX();
	//if(fd < 0) return 0;

	FRAMEINFO_t frameInfo;
	unsigned int frmNo;
	struct timeval tv, tv2;
	printf("Start IPCAM video stream OK!\n");
	int  cnt = 0, fpsCnt = 0, round = 0, lostCnt = 0;
	int outBufSize = 0;
	int outFrmSize = 0;
	int outFrmInfoSize = 0;
	//int bCheckBufWrong;
	int bps = 0;
	//int FlagB=0;
gettimeofday(&tv, NULL);


	while(1)
	{
		if(lostCnt>8) break;
		//Sleep(gSleepMs);
		//ret = avRecvFrameData(avIndex, buf, VIDEO_BUF_SIZE, (char *)&frameInfo, sizeof(FRAMEINFO_t), &frmNo);
		ret = avRecvFrameData2(avIndex, buf, VIDEO_BUF_SIZE, &outBufSize, &outFrmSize, (char *)&frameInfo, sizeof(FRAMEINFO_t), &outFrmInfoSize, &frmNo);

		// show Frame Info at 1st frame
		if(frmNo==1)
		{
			char *format[] = {"MPEG4","H263","H264","MJPEG","UNKNOWN"};
			int idx = 0;
			if(frameInfo.codec_id == MEDIA_CODEC_VIDEO_MPEG4)
				idx = 0;
			else if(frameInfo.codec_id == MEDIA_CODEC_VIDEO_H263)
				idx = 1;
			else if(frameInfo.codec_id == MEDIA_CODEC_VIDEO_H264)
				idx = 2;
			else if(frameInfo.codec_id == MEDIA_CODEC_VIDEO_MJPEG)
				idx = 3;
			else
				idx = 4;
			printf("--- Video Formate: %s ---\n", format[idx]);
		}

		if(ret == AV_ER_DATA_NOREADY)
		{
			//printf("AV_ER_DATA_NOREADY[%d]\n", avIndex);
			Sleep(gSleepMs);
			continue;
		}
		else if(ret == AV_ER_LOSED_THIS_FRAME)
		{
			printf("Lost video frame NO[%d]\n", frmNo);
			lostCnt++;
			//continue;
		}
		else if(ret == AV_ER_INCOMPLETE_FRAME)
		{
			#if 1
			if(outFrmInfoSize > 0)
			printf("Incomplete video frame NO[%d] ReadSize[%d] FrmSize[%d] FrmInfoSize[%u] Codec[%d] Flag[%d]\n", frmNo, outBufSize, outFrmSize, outFrmInfoSize, frameInfo.codec_id, frameInfo.flags);
			else
			printf("Incomplete video frame NO[%d] ReadSize[%d] FrmSize[%d] FrmInfoSize[%u]\n", frmNo, outBufSize, outFrmSize, outFrmInfoSize);
			#endif
			lostCnt++;
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
		}
		else if(ret<0){
			printf("[thread_ReceiveVideo] %d",ret);
			break;
		}
		else
		{
			//if(frameInfo.flags==1) FlagB=1;
			//if(frameInfo.useCount==3){
				CArrayToGoArrayG(buf,ret,frameInfo.flags,num);
			//}
			lostCnt=0;
			//bps += outBufSize;
			//#if 0
			//static int frmCnt = 0;
			//char fn[32];
			//if(frameInfo.flags == IPC_FRAME_FLAG_IFRAME)
			//	sprintf(fn, "I_%03d.bin", frmCnt);
			//else
			//	sprintf(fn, "P_%03d.bin", frmCnt);
			//frmCnt++;
			//FILE *fp = fopen(fn, "wb+");
			//fwrite(buf, 1, outBufSize, fp);
			//fclose(fp);
			//#endif
			//quitBro();
		}

	}
	printf("[thread_ReceiveVideo] thread exit\n");

	//return 0;
}

int Init_Speaker(int sid){
	int talkChannel=4;
	int talkIndex = avServStart(sid, NULL, NULL, 60, 0, talkChannel);
	return talkIndex;
}

int Send_Speaker(int talkindex,char *databuff,unsigned int length,int _audioFrameIndex){
	 FRAMEINFO_t frameInfo;
    memset(&frameInfo, 0, sizeof(FRAMEINFO_t));
    frameInfo.codec_id = 138;
    frameInfo.flags = 2;
//    frameInfo.cam_index = 0;
    frameInfo.onlineNum = 1;
    frameInfo.timestamp = 20 * _audioFrameIndex;
    //_audioFrameIndex ++;
    int ret =avSendAudioData(talkindex, databuff, length, &frameInfo, sizeof(FRAMEINFO_t));

}
typedef struct
{
	unsigned int resolution;
	int usecount;
} SMsgAVIoctrlResolutionMode;

*/
import "C"

import (
	"fmt"
	"github.com/astaxie/beego"
	"math/rand"
	"strconv"
	//"time"
	"unsafe"
)

type P2PCON struct {
	uid         string
	pw          string
	avIndex     int
	SID         int
	linkFlag    bool
	ch          chan *iMessage
	Urlch       chan string
	Broflag     bool
	talkIndex   C.int
	speakerTime int
}

const (
	nRTMPType C.uint = 98765
)

var (
	Localaddr string = "192.168.10.145"
)

func audio_while(avIndex, num C.int) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">>>>>>>>>>>>>>>", err)
			//p.Broflag = false
			//log.Info("%v", err)
		}
		//fmt.Println(">>>>>>>>>>>>>>>")
	}()
	C.thread_ReceiveAudio(avIndex, num)
}
func NewP2PCON(uidsrc string, pw string) *P2PCON {
	return &P2PCON{uid: uidsrc, pw: pw, linkFlag: false, Urlch: make(chan string, 1)}
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

func (p *P2PCON) Write(iotypet uint32, msg string) error {
	iotype := C.uint(iotypet)
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
func (p *P2PCON) Speak(b []byte) {
	ret := C.Send_Speaker(p.talkIndex, (*C.char)(unsafe.Pointer(&b[0])), C.uint(len(b)), C.int(p.speakerTime))
	fmt.Println(">>>>speak>>>>", ret)
	p.speakerTime++
}
func (p *P2PCON) BroadCast() error {
	fmt.Println("#################BEOADCAST##################")
	//ch := make(chan *Message, 1)
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(">>>>>>>>>>>>>>>", err)
			p.Broflag = false
			//log.Info("%v", err)
		}
		fmt.Println(">>>>>>>>>>>>>>>")
	}()
	var num C.int
	ret := C.start_ipcam_stream(C.int(p.avIndex), &num)
	//pic_uid[int(num)] = &picObj{p.uid, false}
	if int(ret) < 0 {
		return fmt.Errorf("start_ipcam_stream error")
	}

	//m := &iMessage{byte(0x09), []byte{0x17, 0, 0, 0, 0, 0x1, 0x4D, 0x00, 0x1E, 0xFF, 0xE1, 0, 0x0A, 0x67, 0x4D, 0x00, 0x1E, 0x95, 0xA8, 0x28, 0x0B, 0xFE, 0x54, 0x1, 0, 0x04, 0x68, 0xEE, 0x3C, 0x80}}
	//videochan <- m
	if int(p.talkIndex) <= 0 {
		p.talkIndex = C.Init_Speaker(C.int(p.SID))
		if int(p.talkIndex) < 0 {
			fmt.Errorf("Init_Speaker error")
		}
	}

	fmt.Println("p.talkIndex=", p.talkIndex)
	p.speakerTime = 0
	ch := listChan[int(num)]
	p.ch = ch
	go broadCast(ch, p.Urlch, "rtmp://"+Localaddr+"/live", p.uid+strconv.Itoa(rand.Intn(100)))
	p.Broflag = true
	fmt.Println(C.int(p.avIndex), num)
	//p.Write(uint32(0x1311), string(1)+string(0)+string(0)+string(0)+string(3)+string(0)+string(0)+string(0))
	go audio_while(C.int(p.avIndex), num)
	C.thread_ReceiveVideo(C.int(p.avIndex), num)

	return nil
}

func (p *P2PCON) StopBro() error {
	if p.ch != nil {
		if p.Broflag == false {
			return nil
		}
		close(p.ch)
		p.Broflag = false
	}
	//C.avServStop(p.talkIndex)
	//fmt.Println("avServStop")
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

	var buf *C.char = (*C.char)(C.malloc(1024))
	defer C.free(unsafe.Pointer(buf))
	//C.memset(buf, C.int(0), C.int(1024))

	var reType C.uint
	ret := C.avRecvIOCtrl(C.int(avIndex), &reType, buf, 1024, C.uint(1000))
	if ret < 0 {
		return "", 0, fmt.Errorf("avSendIOCtrl recive cmd failed[%d]!!", int(ret))
	}
	fmt.Println("stirn-", C.GoString(buf), ret)
	return C.GoString(buf), reType, nil
}

func myclose(avIndex int, SID int) {
	C.avClientStop(C.int(avIndex))
	C.IOTC_Session_Close(C.int(SID))
}
