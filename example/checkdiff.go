package main

import (
	"fmt"
	"strings"

	"github.com/vazrupe/sqlitediff"
)

func main() {
<<<<<<< f963d3c0223ec8da816d44ba42f35f8f65b9f1c7:example/compare.go
	d, err := sqlitediff.Diff("YOUR_BEFORE_DB", "YOUR_BEFORE_DB")
=======
	d, err := sqlitediff.Diff("YOUR_BEFORE_DB", "YOUR_AFTER_DB")
>>>>>>> buffix:example/checkdiff.go
	if err != nil {
		panic(err)
	}

	fmt.Println("AddTable: ", strings.Join(d.AddTables[:], ", "))
	fmt.Println("RemoveTable: ", strings.Join(d.RemoveTables[:], ", "))
	for tableName, tableInfo := range d.ChanageTables {
		fmt.Println("Change: ", tableName)
		fmt.Printf(" - Change Schema (%t)\n", tableInfo.ChangeSchema)
		fmt.Println(" - AddRow: ", tableInfo.AddRows)
		fmt.Println(" - RemoveRow: ", tableInfo.RemoveRows)
		fmt.Println(" - ChangeRow: ", tableInfo.ChangeRows)
	}
}
