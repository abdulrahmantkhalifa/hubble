package hubble

import (
	"github.com/ugorji/go/codec"
	"io"
)

var msgHandle = new(codec.MsgpackHandle)

//Dumps a message to writer. flag it with the given type
func dumps(writer io.Writer, message Message) error {
	//send type byte.
	writer.Write([]byte{uint8(message.GetMessageType())})

	var encoder = codec.NewEncoder(writer, msgHandle)

	return encoder.Encode(message)
}

//Loads a message from a reader, assuming first byte is the message types.
func loads(reader io.Reader) (Message, error) {
	var mtype = make([]byte, 1)
	_, err := reader.Read(mtype)
	if err != nil {
		return nil, err
	}

	msg, err := NewMessage(MessageType(mtype[0]))

	if err != nil {
		return nil, err
	}

	decoder := codec.NewDecoder(reader, msgHandle)

	err = decoder.Decode(msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
