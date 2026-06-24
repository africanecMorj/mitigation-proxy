package pkg

import (
	"errors"
	"strings"
)


type PostgresStartupInfo struct {

	User string

	Database string

	Params map[string]string
}



var (

	ErrNotPostgres = errors.New("not postgres")

)



func ParsePostgresStartup(data []byte) (*PostgresStartupInfo,error){


	if len(data)<8 {
		return nil,ErrTruncated
	}


	length :=
		int(data[0])<<24 |
		int(data[1])<<16 |
		int(data[2])<<8 |
		int(data[3])


	if len(data)<length {
		return nil,ErrTruncated
	}



	// protocol 3.0

	version :=
		int(data[4])<<24 |
		int(data[5])<<16 |
		int(data[6])<<8 |
		int(data[7])


	if version != 196608 {

		return nil,ErrNotPostgres
	}



	info:=&PostgresStartupInfo{
		Params: make(map[string]string),
	}



	payload:=data[8:length]


	parts:=strings.Split(
		string(payload),
		"\x00",
	)



	for i:=0;i+1<len(parts);i+=2 {


		key:=parts[i]

		if key=="" {
			break
		}


		value:=parts[i+1]


		info.Params[key]=value


		switch key {

		case "user":
			info.User=value


		case "database":
			info.Database=value

		}

	}


	return info,nil
}