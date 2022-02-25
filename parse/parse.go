package parse

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"github.com/olemorud/replay-parser/demo"
	"google.golang.org/protobuf/proto"
)

type Frame struct {
	Kind    uint64              `json:"kind"`    // see demo.proto
	Tick    uint64              `json:"tick"`    // time elapsed in replay time
	Size    uint64              `json:"size"`    // size of message in bytes
	Message *demo.CDemoFileInfo `json:"message"` // protobuf encoded message (may be compressed with snappy)
}

// If kind has a certain bit set, the frame message is compressed
// returns true if frame is compressed
func isCompressed(frame *Frame) bool {
	return (frame.Kind&uint64(demo.EDemoCommands_DEM_IsCompressed) != 0)
}

// Checks if file header is correct and returns number the address of the last frame
func First(r *bufio.Reader) (uint64, error) {
	// Check if first 8 bytes of file matches the source 2 replay file header
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, fmt.Errorf("error when reading file: %v", err)
	}
	if string(header) != SIGNATURE {
		return 0, fmt.Errorf("wrong file signature: %v, should be ", SIGNATURE)
	}

	// Read last frame
	var offset uint32
	err := binary.Read(r, binary.LittleEndian, &offset)
	if err != nil {
		return 0, fmt.Errorf("error reading number of frames in replay: %v", err)
	}

	/// remove later
	println("size of demo is", offset, "frames")

	return uint64(offset), nil
}

// Parses the next frame on the reader
func DecodeNextFrame(r *bufio.Reader) (*Frame, error) {
	// Read kind, tick, size and message
	frame := new(Frame)
	var err error

	if frame.Kind, err = binary.ReadUvarint(r); err != nil {
		return nil, fmt.Errorf("error reading frame kind for frame: %v", err)
	}

	if frame.Tick, err = binary.ReadUvarint(r); err != nil {
		return nil, fmt.Errorf("error reading frame tick for frame: %v", err)
	}

	if frame.Size, err = binary.ReadUvarint(r); err != nil {
		return nil, fmt.Errorf("error reading frame size for frame: %v", err)
	}

	message := make([]byte, frame.Size)
	io.ReadFull(r, message)

	frame.Message = new(demo.CDemoFileInfo)

	if isCompressed(frame) {
		decoded, err := snappy.Decode(nil, message)
		if err != nil {
			return nil, fmt.Errorf("error decoding message: %v", err)
		}
		fmt.Println("I got decompressed :O")
		proto.Unmarshal(decoded, frame.Message)
	} else {
		proto.Unmarshal(message, frame.Message)
	}

	return frame, nil
}

// Reads every single frame (probably not very efficient)
func DecodeAllFrames(r *bufio.Reader, frameCount uint64) ([]*Frame, error) {
	replay := make([]*Frame, frameCount)

	// Decode all frames and add them to *Frame slice
	for i := 0; i < int(frameCount); i++ {
		replay[i], _ = DecodeNextFrame(r)
	}

	return replay, nil
}
