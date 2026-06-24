// package inspector

// import (
// 	"io"

// 	"github.com/africanecMorj/mitigation-proxy.git/pkg"
// 	"golang.org/x/sys/unix"
// )


// type Kafka struct {

// 	buf []byte

// 	clientID string

// 	apiKey int16

// 	apiVersion int16
// }



// func (k *Kafka) Read(fd int) (bool,error){


// 	tmp:=make([]byte,4096)


// 	for {


// 		n,err:=unix.Read(fd,tmp)


// 		if err!=nil {


// 			if err==unix.EINTR {
// 				continue
// 			}


// 			if err==unix.EAGAIN {
// 				return false,nil
// 			}


// 			return false,err
// 		}



// 		if n==0 {
// 			return false,io.EOF
// 		}



// 		k.buf=append(k.buf,tmp[:n]...)



// 		// kafka frame length

// 		if len(k.buf)>=4 {


// 			length :=
// 				int(k.buf[0])<<24 |
// 				int(k.buf[1])<<16 |
// 				int(k.buf[2])<<8 |
// 				int(k.buf[3])



// 			if len(k.buf)>=length+4 {


// 				req,err:=pkg.ParseKafkaRequest(
// 					k.buf[:length+4],
// 				)


// 				if err==nil {

// 					k.apiKey=req.APIKey

// 					k.apiVersion=req.APIVersion

// 					k.clientID=req.ClientID

// 				}



// 				return true,nil
// 			}
// 		}
// 	}
// }




// func NewKafka()*Kafka{

// 	return &Kafka{
// 		buf: acquirePreBuf(),
// 	}
// }



// func (k *Kafka) RouteKey() RouteInfo{


// 	return RouteInfo{

// 		ClientID:k.clientID,

// 		Protocol:"kafka",

// 	}

// }



// func (k *Kafka) Data()[]byte{
// 	return k.buf
// }



// func (k *Kafka) Close(){

// 	releasePreBuf(k.buf)

// 	k.buf=nil
// }