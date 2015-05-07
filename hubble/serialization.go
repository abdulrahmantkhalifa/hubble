package hubble

import (
	"io"
	"errors"
	"github.com/ugorji/go/codec"
)

var msgHandle = new(codec.MsgpackHandle)

//Dumps a message to writer. flag it with the given type
func Dumps(writer io.Writer, mtype uint8, message interface{}) error {
	//send type byte.
	writer.Write([]byte{mtype})

	var encoder = codec.NewEncoder(writer, msgHandle)

	return encoder.Encode(message)
}

//Loads a message from a reader, assuming first byte is the message types.
func Loads(reader io.Reader) (uint8, interface{}, error) {
	var mtype = make([]byte, 1)
	_, err := reader.Read(mtype)
	if err != nil {
		return 0, nil, err
	}

	initiator, ok := MessageHeaderTypes[mtype[0]]
	if !ok {
		return mtype[0], nil, errors.New("Invalid mtype")
	}

	decoder := codec.NewDecoder(reader, msgHandle)
	var value = initiator()
	//var loaded int
	
	derr := decoder.Decode(value)
	if derr != nil {
		return mtype[0], nil, derr
	}

	return mtype[0], value, nil
}
