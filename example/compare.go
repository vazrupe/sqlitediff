package main

import (
	"fmt"
	"strings"

	"github.com/vazrupe/sqlitediff"
)

func main() {
	d, err := sqlitediff.Diff("master_be.mdb", "master_af.mdb")
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
