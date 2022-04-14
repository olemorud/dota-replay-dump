package parse

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/olemorud/replay-parser/demo"
	"google.golang.org/protobuf/proto"
)

type Frame struct {
	Kind            demo.EDemoCommands `json:"kind"`    // see demo.proto
	Tick            uint64             `json:"tick"`    // time elapsed in replay time
	Message         proto.Message      `json:"message"` // protobuf encoded message (may be compressed with snappy)
	HasEmbeddedData bool
}

// Return true if the frame message is snappy compressed
func (frame *Frame) isCompressed() bool {
	return (frame.Kind&demo.EDemoCommands_DEM_IsCompressed != 0)
}

// Checks if file header is correct and returns number the address of the last frame
func First(r *bufio.Reader) (uint64, error) {
	// Check if first 8 bytes of file is PBDEMS2\0
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, fmt.Errorf("error when reading file: %v", err)
	}
	if string(header) != SOURCE2_SIGN {
		return 0, fmt.Errorf("wrong file signature: %v, should be ", SOURCE2_SIGN)
	}

	// Read offset to last frame
	var offset uint32
	err := binary.Read(r, binary.LittleEndian, &offset)
	if err != nil {
		return 0, fmt.Errorf("error reading number of frames in replay: %v", err)
	}

	return uint64(offset), nil
}

// Parses the next frame on the reader
func DecodeNextFrame(r *bufio.Reader) (*Frame, error) {

	// Read command
	c, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("error reading frame command: %v", err)
	}
	command := demo.EDemoCommands(c)

	// Read tick
	tick, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("error reading frame tick: %v", err)
	}
	if tick == 0xFFFFFFFF {
		tick = 0
	}

	// Read size
	size, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("error reading frame size: %v", err)
	}

	// Instanciate frame with data
	frame := &Frame{
		Kind:            command,
		Tick:            tick,
		Message:         nil,
		HasEmbeddedData: false,
	}
	frame.setMessageType()

	// Debug
	if !frame.HasEmbeddedData {
		fmt.Printf(`
		kind: %d
		tick: %d
		size: %d
		`, command & ^demo.EDemoCommands_DEM_IsCompressed, tick, size)
	}

	if frame.HasEmbeddedData {
		r.Discard(int(size))
		return nil, nil
	}

	// Read message
	message := make([]byte, size)
	io.ReadFull(r, message)

	if frame.isCompressed() {
		decoded, err := snappy.Decode(nil, message)
		if err != nil {
			return nil, fmt.Errorf("error decoding message: %v", err)
		}
		proto.Unmarshal(decoded, frame.Message)
	} else {
		proto.Unmarshal(message, frame.Message)
	}

	return frame, nil
}

// I hate reflection
func (f *Frame) setMessageType() {

	switch f.Kind & ^demo.EDemoCommands_DEM_IsCompressed {
	default:
		f.Message = nil
	// case demo.EDemoCommands_DEM_Error:
	// 	f.Message = &demo.CDemoError{}
	case demo.EDemoCommands_DEM_Stop:
		f.Message = &demo.CDemoStop{}
	case demo.EDemoCommands_DEM_FileHeader:
		f.Message = &demo.CDemoFileHeader{}
	case demo.EDemoCommands_DEM_FileInfo:
		f.Message = &demo.CDemoFileInfo{}
	case demo.EDemoCommands_DEM_SyncTick:
		f.Message = &demo.CDemoSyncTick{}
	case demo.EDemoCommands_DEM_SendTables:
		f.Message = &demo.CDemoSendTables{}
		f.HasEmbeddedData = true
	case demo.EDemoCommands_DEM_ClassInfo:
		f.Message = &demo.CDemoClassInfo{}
	case demo.EDemoCommands_DEM_StringTables:
		f.Message = &demo.CDemoStringTables{}
	case demo.EDemoCommands_DEM_Packet:
		f.Message = &demo.CDemoPacket{}
		f.HasEmbeddedData = true
	case demo.EDemoCommands_DEM_SignonPacket:
		f.Message = nil //&demo.CDemoSignonPacket{}
		f.HasEmbeddedData = true
	case demo.EDemoCommands_DEM_ConsoleCmd:
		f.Message = &demo.CDemoConsoleCmd{}
	case demo.EDemoCommands_DEM_CustomData:
		f.Message = &demo.CDemoCustomData{}
	case demo.EDemoCommands_DEM_CustomDataCallbacks:
		f.Message = &demo.CDemoCustomDataCallbacks{}
	case demo.EDemoCommands_DEM_UserCmd:
		f.Message = &demo.CDemoUserCmd{}
	case demo.EDemoCommands_DEM_FullPacket:
		f.Message = &demo.CDemoFullPacket{}
		f.HasEmbeddedData = true
	case demo.EDemoCommands_DEM_SaveGame:
		f.Message = &demo.CDemoSaveGame{}
	case demo.EDemoCommands_DEM_SpawnGroups:
		f.Message = &demo.CDemoSpawnGroups{}
	}

}

// Following code is stolen from binary package
const MaxVarintLen32 = 5

func ReadUvarint32(r io.ByteReader) (uint32, error) {
	var x uint32
	var s uint
	for i := 0; i < MaxVarintLen32; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return x, err
		}
		if b < 0x80 {
			if i == MaxVarintLen32-1 && b > 1 {
				return x, errors.New("binary: varint overflows a 32-bit integer")
			}
			return x | uint32(b)<<s, nil
		}
		x |= uint32(b&0x7f) << s
		s += 7
	}
	return x, errors.New("binary: varint overflows a 32-bit integer")
}
