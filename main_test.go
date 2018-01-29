package main

import (
	"log"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func TestJsonUnmarshal(t *testing.T) {
	_, err := LoadJSON("testdata/roles.json")
	if err != nil {
		t.Fatal(err)
	}
}

func mkXslx(t *testing.T, filename string) {
	xslx := excelize.NewFile()
	// Sheet1 for project roles, auto-generated
	// Header
	xslx.SetCellValue("Sheet1", "A1", "Project roles")
	xslx.SetCellValue("Sheet1", "B1", "Jenkins permissions")
	xslx.SetCellValue("Sheet1", "C1", "Pattern")

	// Add build role
	xslx.SetCellValue("Sheet1", "A2", "build")
	xslx.SetCellValue("Sheet1", "B2",
		"hudson.model.Item.Discover\t"+
			"hudson.model.Item.Build\t"+
			"hudson.model.Item.Workspace")
	xslx.SetCellValue("Sheet1", "C2", "build-.*")

	// Add deployment role
	xslx.SetCellValue("Sheet1", "A3", "deployment")
	xslx.SetCellValue("Sheet1", "B3", "hudson.model.Item.Discover")
	xslx.SetCellValue("Sheet1", "C3", "deploy-.*")

	// Assigned users
	xslx.NewSheet("Sheet2")
	xslx.SetCellValue("Sheet2", "A2", "build")
	xslx.SetCellValue("Sheet2", "B2", "testuser1")
	xslx.SetCellValue("Sheet2", "C2", "testuser2")
	xslx.SetCellValue("Sheet2", "D2", "testuser3")

	xslx.SetCellValue("Sheet2", "A3", "deployment")
	xslx.SetCellValue("Sheet2", "B3", "deployuser1")
	xslx.SetCellValue("Sheet2", "C3", "deployuser2")

	err := xslx.SaveAs(filename)
	die(err)
	log.Printf("wrote %s\n", filename)
}

func TestGenXslx(t *testing.T) {
	mkXslx(t, "testdata/gen1.xslx")
}

func TestXslxImport(t *testing.T) {
	mkXslx(t, "testdata/gen2.xslx")
	roles, err := LoadXslx("testdata/gen2.xslx")
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 2 {
		t.Fatalf("want 2 but got %d\n", len(roles))
	}
}
