// Evrista (SRC Giants) file parser (format reverse-engineered).
package evrista

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"unsafe"
)

// File contains all information from Evrista database.
type File struct {
	// Comment up to 52 ASCII symbols.
	Comment string

	// Data series.
	Series []Series
}

// Series describe one data column.
type Series struct {
	// Name up to 11 ASCII symbols.
	Name string

	// Values, some of which may be math.NaN().
	Values []float64

	dataType dataType
}

type dataType byte

const (
	dtDouble dataType = 0x06
	dtFloat           = 0x05
)

var nanFloat = *(*float32)(unsafe.Pointer(&[4]byte{'*', '*', '*', '*'}))
var nanDouble = *(*float64)(unsafe.Pointer(&[8]byte{'*', '*', '*', '*', '*', '*', '*', '*'}))

// Parse Evrista database (*.gnt) into a File structure.
func Parse(in io.Reader) (*File, error) {
	var file File

	comment := make([]byte, 52)
	if err := binary.Read(in, binary.LittleEndian, comment); err != nil {
		return nil, err
	}
	file.Comment = makeString(comment)

	var count byte
	if err := binary.Read(in, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	file.Series = make([]Series, count)

	if err := skip(in, 16); err != nil {
		return nil, err
	}

	for i := 0; i < len(file.Series); i++ {
		series := &file.Series[i]

		name := make([]byte, 11)
		if err := binary.Read(in, binary.LittleEndian, name); err != nil {
			return nil, err
		}
		series.Name = makeString(name)

		var dataType dataType
		if err := binary.Read(in, binary.LittleEndian, &dataType); err != nil {
			return nil, err
		}

		switch dataType {
		case dtDouble, dtFloat:
			series.dataType = dataType
		default:
			return nil, fmt.Errorf("unknown series '%s' data type %v", series.Name, dataType)
		}

		if err := skip(in, 2); err != nil {
			return nil, err
		}

		var size uint32
		if err := binary.Read(in, binary.LittleEndian, &size); err != nil {
			return nil, err
		}
		series.Values = make([]float64, size)

		if err := skip(in, 87); err != nil {
			return nil, err
		}
	}

	for i := range file.Series {
		series := &file.Series[i]

		switch series.dataType {
		case dtDouble:
			if err := binary.Read(in, binary.LittleEndian, series.Values); err != nil {
				return nil, err
			}
			for j := 0; j < len(series.Values); j++ {
				if series.Values[j] == nanDouble {
					series.Values[j] = math.NaN()
				}
			}
		case dtFloat:
			values := make([]float32, len(series.Values))
			if err := binary.Read(in, binary.LittleEndian, values); err != nil {
				return nil, err
			}
			for j := range values {
				if values[j] == nanFloat {
					series.Values[j] = math.NaN()
				} else {
					series.Values[j] = float64(values[j])
				}
			}
		default:
			panic("unsupported data type") // should be validated earlier
		}
	}

	return &file, nil
}

func makeString(bytes []byte) string {
	var i int
	for i = 0; i < len(bytes) && bytes[i] != 0; i++ {}
	return string(bytes[:i])
}

func skip(in io.Reader, count int64) error {
	switch r := in.(type) {
	case io.Seeker:
		_, err := r.Seek(count, io.SeekCurrent)
		return err
	default:
		_, err := io.CopyN(ioutil.Discard, in, count)
		return err
	}
}
