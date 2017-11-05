package xpt

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

// List of valid header parts
var (
	headerError = "incorrect header record format"
	libHeader   = "HEADER RECORD*******LIBRARY HEADER RECORD" +
		"!!!!!!!000000000000000000000000000000"
	EXPECTED_LIB_HEADER = libHeader + "  "
	SAS_HEADER          = "SAS                                      " +
		"                                       "
	MODIFIED_HEADER = " 21JAN08:13:32:41                        " +
		"                                       "
	memHeader      = "HEADER RECORD*******MEMBER  HEADER RECORD!!!!!!!"
	MEM_HEADER     = memHeader + "000000000000000001600000000140  "
	dscrptr_header = "HEADER RECORD*******DSCRPTR HEADER RECORD!!!!!!!"
	DSCRPTR_HEADER = dscrptr_header + "                                "
	DATASET_HEADER = "        AE                                      " +
		"                                "
	namestr_header = "HEADER RECORD*******NAMESTR HEADER RECORD!!!!!!!"
	NAMESTR_HEADER = namestr_header + "000000001200000000000000000000  "
	obser_header   = "HEADER RECORD*******OBS     HEADER RECORD!!!!!!!"
	OBSER_HEADER   = obser_header + "000000000000000000000000000000  "
)

func TestReadInvalidHeader(t *testing.T) {
	testFailureMsg := "%v\n Expected error: \n'%v' \nActual Error: \n'%v'"

	strReader := strings.NewReader("Test")
	reader := bufio.NewReader(strReader)

	_, err := NewXptReader(reader)
	expectedErr := fmt.Sprintf(
		"%v: read 4 out of 640 bytes of library header", headerError)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, "invalid length", expectedErr, err.Error())
	}

	// Test invalid lib header
	badHeader := strings.Replace(EXPECTED_LIB_HEADER, "H", "P", -1)
	header := badHeader + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)

	expectedErr = fmt.Sprintf("%v: expected to contain: '%v' but got: '%v'",
		headerError, libHeader, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, "", expectedErr, err.Error())
	}

	// Test invalid sas header
	badHeader = strings.Replace(SAS_HEADER, "S", "B", -1)
	header = EXPECTED_LIB_HEADER + badHeader + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)

	expectedErr = fmt.Sprintf("%v: expected first real header record got: '%v'",
		headerError, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, "", expectedErr, err.Error())
	}

	// Test invalid modified header
	// Maybe add a date format validation?

	// Test invalid member header
	badHeader = strings.Replace(MEM_HEADER, "H", "B", -1)
	header = EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		badHeader + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)

	expectedErr = fmt.Sprintf("%v: expected to contain: '%v' but got: '%v'",
		headerError, memHeader, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, "", expectedErr, err.Error())
	}

	testName := "Test invalid descriptor header"
	badHeader = strings.Replace(DSCRPTR_HEADER, "H", "B", -1)
	header = EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + badHeader + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)

	expectedErr = fmt.Sprintf("%v: expected to contain: '%v' but got: '%v'",
		headerError, dscrptr_header, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, testName, expectedErr, err.Error())
	}

	// Test invalid dataset header

	// Test modified header again?

	testName = "Test invalid namestr header"
	badHeader = strings.Replace(NAMESTR_HEADER, "H", "B", -1)
	header = EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + badHeader
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)
	expectedErr = fmt.Sprintf("%v: expected to contain: '%v' but got: '%v'",
		headerError, namestr_header, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, testName, expectedErr, err.Error())
	}

	testName = "Test failure with missing variable records"
	header = EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	strReader = strings.NewReader(header)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)
	expectedErr = fmt.Sprintf(
		"Incomplete or empty variable record %v out of %v bytes", 0, 3500)
	if err == nil {
		t.Errorf(testFailureMsg, testName, expectedErr, err.Error())
	}

	testName = "Test with invalid observation header"
	badHeader = strings.Replace(OBSER_HEADER, "H", "B", -1)
	header = EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	paddingAndVariableRecords := fmt.Sprintf("%1760s", "") + badHeader
	strReader = strings.NewReader(header + paddingAndVariableRecords)
	reader = bufio.NewReader(strReader)
	_, err = NewXptReader(reader)
	expectedErr = fmt.Sprintf("expected to contain: '%v' but got: '%v'",
		obser_header, badHeader)
	if err.Error() != expectedErr {
		t.Errorf(testFailureMsg, testName, expectedErr, err.Error())
	}
}

func TestReadValidHeader(t *testing.T) {
	testName := "Test valid header"
	// (12(variables) * 140(variableRecordSize)) + 80 padding = 1760
	header := EXPECTED_LIB_HEADER + SAS_HEADER + MODIFIED_HEADER +
		MEM_HEADER + DSCRPTR_HEADER + DATASET_HEADER +
		MODIFIED_HEADER + NAMESTR_HEADER
	paddingAndVariableRecords := fmt.Sprintf("%1760s", "") + OBSER_HEADER
	strReader := strings.NewReader(header + paddingAndVariableRecords)
	reader := bufio.NewReader(strReader)
	xptReader, err := NewXptReader(reader)
	if err != nil {
		t.Fatalf("%v\nvalid header failed to be parsed \n%v",
			testName, err.Error())
	}

	// Variable count is extracted from NAMESTR_HEADER
	expectedVariableCount := 12
	if xptReader.header.variableCount != expectedVariableCount {
		t.Errorf("%v\nincorrectly read variable count as %v should be %v",
			testName, xptReader.header.variableCount, expectedVariableCount)
	}

	// Variable record size is extracted from the MEMBER_HEADER
	expectedVariableRecordSize := 140
	if xptReader.header.variableRecordSize != expectedVariableRecordSize {
		t.Errorf("%v\nincorrectly read variable record length actual %v expected %v",
			testName, xptReader.header.variableRecordSize, expectedVariableRecordSize)
	}

	//expectedXptCreated := ""
	//if xptReader.header.created != expectedXptCreated {
	//	t.Errorf("%v\nIncorrectly read xpt created date actual %v expected %v",
	//		testName, xptReader.header.created, expectedXptCreated)
	//}

	// Test that the dataset name is read from the dataset header record
	expectedDataset := "AE"
	if xptReader.header.dataset != expectedDataset {
		t.Errorf("%v\nIncorrectly read xpt dataset name actual %v expected %v",
			testName, xptReader.header.dataset, expectedDataset)
	}

	expectedModifiedDate := "21JAN08:13:32:41"
	if xptReader.header.modified != expectedModifiedDate {
		t.Errorf("%v\nIncorrectly read xpt modified date actual %v expected %v",
			testName, xptReader.header.modified, expectedModifiedDate)
	}
}

func TestReadVariables(t *testing.T) {

}
