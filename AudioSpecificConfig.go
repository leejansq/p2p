package P2P

import ()

type adts struct {
	syncword                 int16
	ID                       byte
	layer                    byte
	protection_absent        byte
	profile                  byte
	sampling_frequency_index byte
	private_bit              byte
	channel_configuration    byte
	original_copy            byte
	home                     byte
}

type myError struct {
	content string
	Err     error
}

func (s *myError) Error() string {
	return "err:" + s.content + s.Err.Error()
}
func AdtsToConfig(b []byte) ([]byte, error) {
	if len(b) < 7 {
		e := &myError{content: "len < 7"}
		return nil, e
	}
	profile := ((b[2] & 0xc0) >> 6) + 1
	sample_rate := (b[2] & 0x3c) >> 2
	channel := ((b[2] & 0x1) << 2) | ((b[3] & 0xc0) >> 6)
	config1, config2 := (profile<<3)|((sample_rate&0xe)>>1), ((sample_rate&0x1)<<7)|(channel<<3)
	//config2 := ((sample_rate&0x1)<<7)|(channel<<3)
	su := []byte{}
	su = append(su, config1, config2)
	return su, nil

}
