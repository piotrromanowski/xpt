package xpt

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/piotrromanowski/xpt/math"
)

// A Reader parses the header records of a SAS XPT File and enables a user
// to read the observation records from it
// The structure of a valid SAS XPT File is defined in the following spec:
// https://support.sas.com/techsup/technote/ts140.pdf
type Reader struct {
	r            *bufio.Reader
	currentBytes []byte
	header       Header
	Variables    []Variable
	recordLength uint16
}

type Variable struct {
	Name      string
	isNumeric bool
	position  uint16
	length    uint16
	number    uint16
	label     string
}

type Header struct {
	sasType            string
	dataset            string
	sasLib             string
	sasVer             string
	sasOs              string
	blanks             string
	created            string
	modified           string
	variableCount      int
	variableRecordSize int
}

// NewXptReader
// Returns a Reader that reads from r, or an error
func NewXptReader(r *bufio.Reader) (reader *Reader, err error) {
	reader = &Reader{
		r: r,
	}
	if err = readHeader(reader); err != nil {
		return nil, fmt.Errorf("incorrect header record format: %v", err)
	}
	if err = readVariables(reader); err != nil {
		return nil, err
	}
	return
}

var (
	recordLength              = 80
	XPT_LIBRARY_HEADER_RECORD = "HEADER RECORD*******LIBRARY HEADER RECORD" +
		"!!!!!!!000000000000000000000000000000"
	MEMBER_HEADER_RECORD  = "HEADER RECORD*******MEMBER  HEADER RECORD!!!!!!!"
	DSCRPTR_HEADER_RECORD = "HEADER RECORD*******DSCRPTR HEADER RECORD!!!!!!!"
	NAME_STR_HEADER       = "HEADER RECORD*******NAMESTR HEADER RECORD!!!!!!!"
	OBS_HEADER            = "HEADER RECORD*******OBS     HEADER RECORD!!!!!!!"
)

func checkHeader(buf []byte, header string) (err error) {
	if !bytes.Contains(buf, []byte(header)) {
		return fmt.Errorf("expected to contain: '%v' but got: '%v'",
			header, string(buf))
	}
	return
}

func readHeader(reader *Reader) error {
	// Pre-check length of header record once so that we don't have to do as
	// much error checking for amount of bytes read for each part of header record
	readBytes, err := reader.r.Peek(recordLength * 8)
	if err != nil {
		return fmt.Errorf("read %v out of %v bytes of library header",
			len(readBytes), recordLength*8)
	}
	buf := make([]byte, recordLength)

	// line 1
	_, _ = reader.r.Read(buf)
	if err = checkHeader(buf, XPT_LIBRARY_HEADER_RECORD); err != nil {
		return err
	}
	// line 2
	_, _ = reader.r.Read(buf)
	if !bytes.Contains(buf, []byte("SAS")) {
		return fmt.Errorf("expected first real header record got: '%v'",
			string(buf))
	}
	reader.header.sasVer = string(buf[24:32])
	reader.header.sasOs = string(buf[32:40])
	reader.header.created = string(buf[74:])
	// line 3
	_, _ = reader.r.Read(buf)
	reader.header.modified = strings.TrimSpace(string(buf))
	// line 4
	_, _ = reader.r.Read(buf)
	if err = checkHeader(buf, MEMBER_HEADER_RECORD); err != nil {
		return err
	}
	variableRecordSize := buf[48+26 : 48+26+4]
	variableRecordSizeInt, err := strconv.Atoi(string(variableRecordSize))
	if err != nil {
		// TODO: Missing test case
		return fmt.Errorf("error getting variable record length: %v", err.Error())
	}
	reader.header.variableRecordSize = variableRecordSizeInt

	// line 5
	_, _ = reader.r.Read(buf)
	if err = checkHeader(buf, DSCRPTR_HEADER_RECORD); err != nil {
		return err
	}
	// line 6
	_, _ = reader.r.Read(buf)
	reader.header.dataset = strings.TrimSpace(string(buf[8:16]))
	// line 7
	_, _ = reader.r.Read(buf)
	reader.header.modified = strings.TrimSpace(string(buf[0:17]))
	// line 8
	_, _ = reader.r.Read(buf)
	if err = checkHeader(buf, NAME_STR_HEADER); err != nil {
		return err
	}

	varCount := string(buf[54:58])
	if reader.header.variableCount, err = strconv.Atoi(varCount); err != nil {
		// TODO: Missing test case
		return fmt.Errorf("error getting variable count: %v", err)
	}
	return nil
}

func readVariables(reader *Reader) error {
	readBytes, err := reader.r.Peek(reader.header.variableRecordSize *
		reader.header.variableCount)
	if err != nil {
		return fmt.Errorf("Incomplete or empty variable record %v out of %v bytes",
			len(readBytes),
			reader.header.variableRecordSize*reader.header.variableCount)
	}

	variables := []Variable{}
	currentVar := make([]byte, reader.header.variableRecordSize)

	for i := 0; i < reader.header.variableCount; i++ {
		count, err := io.ReadFull(reader.r, currentVar)
		if err != nil {
			return fmt.Errorf("Problem reading variable record %v", err.Error())
		}
		if count < reader.header.variableRecordSize {
			fmt.Printf("Possible error reading variable %v  count %v \n",
				string(currentVar), count)
		}

		varibleLen := binary.BigEndian.Uint16(currentVar[4:6])
		variables = append(variables, Variable{
			Name:      strings.TrimSpace(string(currentVar[8:16])),
			isNumeric: binary.BigEndian.Uint16(currentVar[0:2]) == 1,
			number:    binary.BigEndian.Uint16(currentVar[6:8]),
			position:  binary.BigEndian.Uint16(currentVar[86:88]),
			label:     string(currentVar[16:]),
			length:    varibleLen,
		})
		reader.recordLength += varibleLen
	}
	reader.Variables = variables

	// Read Padding after variables
	padding := 80 - ((reader.header.variableCount *
		reader.header.variableRecordSize) % 80)
	rest := make([]byte, padding)
	if _, err = reader.r.Read(rest); err != nil {
		return fmt.Errorf("missing padding: %v", err)
	}

	// Read Observation header directly the preceeds the actual records
	observationHeader := make([]byte, 80)
	_, _ = io.ReadFull(reader.r, observationHeader)
	if err = checkHeader(observationHeader, OBS_HEADER); err != nil {
		return err
	}
	return nil
}

type ObservationRecord struct {
	raw                []byte
	GetDataForVariable func(varName string) (string, error)
}

// Read reads the next record in the xpt file
// It returns an ObservationRecord and error
// When the reader reaches the end of file it will return a nil ObservationRecord
// and io.EOF error
func (reader *Reader) Read() (*ObservationRecord, error) {
	record := make([]byte, reader.recordLength)
	if _, err := io.ReadFull(reader.r, record); err != nil {
		return nil, err
	}

	observationRecord := &ObservationRecord{
		raw: record, // raw is never used
	}

	getVariable := func(varName string) *Variable {
		for _, v := range reader.Variables {
			if v.Name == varName {
				return &v
			}
		}
		return nil
	}

	// Unsure if this is the correct way to design this api, and/or if
	// dynamically assigning methods is the cool thing to do
	observationRecord.GetDataForVariable = func(varName string) (string, error) {
		// Should this function take in a variable name or a variable? I think
		// variable would make more sense since it would be faster (Not having to
		// look up var) and no validation
		variable := getVariable(varName)
		if variable == nil {
			return "", fmt.Errorf("variable does not exist: %v", varName)
		}

		// Check length of buffer to see if position:length exists
		if variable.position+variable.length > reader.recordLength {
			return "", fmt.Errorf(
				"variable position out of bounds of record '%v'", varName)
		}

		s := string(record[variable.position : variable.position+variable.length])

		if variable.isNumeric {
			num := math.IbmToIEEE(
				record[variable.position : variable.position+variable.length])
			s = strconv.FormatFloat(num, 'f', 6, 64)
		}

		return s, nil
	}

	return observationRecord, nil
}
