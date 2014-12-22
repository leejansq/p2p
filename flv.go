package P2P

import "C"
import (
	"encoding/binary"
	"fmt"
	"unsafe"
)

type iMessage struct {
	tagtype byte
	data    []byte
}

var (
	once     func(C.int)
	listChan []chan *iMessage = []chan *iMessage{}
)

func factory() func(C.int) {
	flag := true
	return func(i C.int) {
		if flag == true {
			flag = false
			iMess := &iMessage{}
			iMess.tagtype = 0x09
			iMess.data = []byte{0x17, 0, 0, 0, 0, 0x1, 0x4D, 0x00, 0x1E, 0xFF, 0xE1, 0, 0x0A, 0x67, 0x4D, 0x00, 0x1E, 0x95, 0xA8, 0x28, 0x0B, 0xFE, 0x54, 0x1, 0, 0x04, 0x68, 0xEE, 0x3C, 0x80}
			listChan[int(i)] <- iMess
		}
	}
}

func Int64ToBytes(i uint32) []byte {
	var buf = make([]byte, 4)
	//binary.(buf, uint64(i))
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}

func CArrayToGoArray(cArray unsafe.Pointer, size uint32, flag int) (goArray []byte) {
	//goArray = make([]byte, size)
	//flag := *(*int)(unsafe.Pointer(C.keyFlag))

	//flag := int(C.frameInfo.flags)
	//fmt.Println("TTTTTTTTTTTTTTTTTTTT", httpCount, "==", flag)
	//temp := []byte{0, 0, 0, 10}
	p := uintptr(cArray)
	flag1 := 0
	flag2 := 0
	//mp := uint32(0)
	if flag == 1 {
		mp := uint32(5)
		temp := Int64ToBytes(size - 4)
		goArray = []byte{23, 1, 0, 0, 0}
		for i := uint32(0); i < size; i++ {
			if i < 4 {
				goArray = append(goArray, 0)
				p += 1
				continue
			}
			j := *(*byte)(unsafe.Pointer(p))

			if flag1 >= 3 {
				if j == 1 {
					temp = Int64ToBytes(i + 5 - 3 - (mp + 4))
					goArray[mp] = temp[0]
					goArray[mp+1] = temp[0+1]
					goArray[mp+2] = temp[0+2]
					goArray[mp+3] = temp[0+3]
					flag2++
					mp = i + 5 - 3
				}

				//n:=i-3
				//s:=Int64ToBytes(n-mp-4)

			}

			if j == 0 {
				flag1++
			} else {
				flag1 = 0
			}

			goArray = append(goArray, j)
			if flag2 >= 1 && i == (size-1) {
				temp = Int64ToBytes(size + 5 - (mp + 4))
				goArray[mp] = temp[0]
				goArray[mp+1] = temp[0+1]
				goArray[mp+2] = temp[0+2]
				goArray[mp+3] = temp[0+3]
			}
			p += 1

		}
	} else {
		mp := uint32(5)
		goArray = []byte{39, 1, 0, 0, 0}
		temp := Int64ToBytes(size - 4)
		for i := uint32(0); i < size; i++ {

			j := *(*byte)(unsafe.Pointer(p))

			if i < 4 {
				goArray = append(goArray, temp[i])
				p += 1
				continue
			}
			//else {
			//goArray = append(goArray, j)
			//}
			//goArray = append(goArray, 11)
			if flag1 >= 3 {
				if j == 1 {
					temp = Int64ToBytes(i + 5 - 3 - (mp + 4))
					goArray[mp] = temp[0]
					goArray[mp+1] = temp[0+1]
					goArray[mp+2] = temp[0+2]
					goArray[mp+3] = temp[0+3]
					flag2++
					mp = i + 5 - 3
				}

				//n:=i-3
				//s:=Int64ToBytes(n-mp-4)

			}

			if j == 0 {
				flag1++
			} else {
				flag1 = 0
			}

			goArray = append(goArray, j)
			if flag2 >= 1 && i == (size-1) {
				temp = Int64ToBytes(size + 5 - (mp + 4))
				goArray[mp] = temp[0]
				goArray[mp+1] = temp[0+1]
				goArray[mp+2] = temp[0+2]
				goArray[mp+3] = temp[0+3]
			}
			p += 1

		}
	}

	return
}

//export addPubchan
func addPubchan() C.int {
	ch := make(chan *iMessage, 3)
	listChan = append(listChan, ch)
	once = factory()
	return C.int(len(listChan) - 1)
}

//export CArrayToGoArrayG
func CArrayToGoArrayG(cArray unsafe.Pointer, size C.int, flag C.int, num C.int) {
	if int(size) < 20 {
		return
	}
	once(num)
	iMess := &iMessage{}
	iMess.tagtype = 0x09
	iMess.data = CArrayToGoArray(cArray, uint32(size), int(flag))
	fmt.Println(iMess.data[:20])
	listChan[int(num)] <- iMess
}

//export HotAudio
func HotAudio(cArray unsafe.Pointer, size C.int, flag C.int, num C.int) {
	if int(size) < 7 {
		return
	}
	bArr := C.GoBytes(cArray, size)
	iMess := &iMessage{}
	iMess.tagtype = 0x08
	if int(flag) == 1 {
		aArr, _ := AdtsToConfig(bArr[0:7])
		hed := []byte{0xAC, 0x0}
		iMess.data = append(hed, aArr...)
	} else {
		//FAAD.DecodeFAAD(bArr)
		hed := []byte{0xAC, 0x01}
		iMess.data = append(hed, bArr[7:]...)
	}
	listChan[int(num)] <- iMess
}
