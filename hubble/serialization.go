package hubble

import (
	"io"
	"errors"
	"github.com/ugorji/go/codec"
)

var msgHandle = new(codec.MsgpackHandle)

//Dumps a message to writer. flag it with the given type
func dumps(writer io.Writer, mtype uint8, flags uint8, message interface{}) error {
	//send type byte.
	writer.Write([]byte{mtype, flags})

	var encoder = codec.NewEncoder(writer, msgHandle)

	return encoder.Encode(message)
}

//Loads a message from a reader, assuming first byte is the message types.
func loads(reader io.Reader) (uint8, uint8, interface{}, error) {
	var header = make([]byte, 2)
	_, err := reader.Read(header)
	if err != nil {
		return 0, 0, nil, err
	}
	
	var mtype uint8 = header[0]
	var flags uint8 = header[1]

	initiator, ok := MessageHeaderTypes[mtype]
	if !ok {
		return mtype, flags, nil, errors.New("Invalid mtype")
	}

	decoder := codec.NewDecoder(reader, msgHandle)
	var value = initiator()
	//var loaded int
	
	derr := decoder.Decode(value)
	if derr != nil {
		return mtype, flags, nil, derr
	}

	return mtype, flags, value, nil
}
