package pkg


import (
	"errors"
)


var ErrNotKafka = errors.New("not kafka")



type KafkaRequest struct {

	APIKey int16

	APIVersion int16

	CorrelationID int32

	ClientID string
}



func ParseKafkaRequest(data []byte)(*KafkaRequest,error){


	if len(data)<14 {
		return nil,ErrTruncated
	}



	length:=int32(
		uint32(data[0])<<24 |
		uint32(data[1])<<16 |
		uint32(data[2])<<8 |
		uint32(data[3]),
	)



	if int(length)+4 > len(data){
		return nil,ErrTruncated
	}



	r:=&KafkaRequest{}


	r.APIKey =
		int16(data[4])<<8 |
		int16(data[5])



	r.APIVersion =
		int16(data[6])<<8 |
		int16(data[7])



	r.CorrelationID =
		int32(data[8])<<24 |
		int32(data[9])<<16 |
		int32(data[10])<<8 |
		int32(data[11])



	// client_id string

	offset:=12


	if len(data)<offset+2 {
		return nil,ErrTruncated
	}


	l:=
		int(data[offset])<<8 |
		int(data[offset+1])


	offset+=2



	if len(data)<offset+l {
		return nil,ErrTruncated
	}



	r.ClientID=
		string(data[offset:offset+l])



	return r,nil
}