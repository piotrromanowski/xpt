# xpt
![coverage](https://svg-badge.appspot.com/badge/coverage/71.4%25?color=3a3)

A go library to read SAS TRANSPORT (XPORT) FORMAT files.

It was originally inspired by https://github.com/selik/xport, but
mostly developed as a learning tool.

Warning: This library is not completely functional. The ibmToIEEE()
parser requires work, amongst other things. Please use at your
own risk and report any bugs as github issues.

#### Reader Usage:
  ```go
	  file, err := os.Open("/path/to/my_file.xpt")
	  if err != nil {
	  	fmt.Printf("\nFile Not Found %v", err)
	  }
    defer file.Close()

	  xptReader, err := xpt.NewXptReader(bufio.NewReader(file))
	  if err != nil {
	  	log.Fatal(err)
	  }

	  varNames := []string{}
	  for _, v := range xptReader.Variables {
	  	varNames = append(varNames, v.Name)
	  }

	  var readError error
	  for readError == nil {
		  record := []string{}

		  r, readError := xptReader.Read()
		  if readError != nil {
		  	break
		  }
		  for _, recordVariable := range xptReader.Variables {
		  	varData, err := r.GetDataForVariable(recordVariable.Name)
		  	if err != nil {
		  		log.Printf("could not read variable: %v", err.Error())
		  		continue
		  	}
		  	record = append(record, varData)
		  }
      fmt.Println(record)
    }
  ```

#### Writer Usage:
  TODO
