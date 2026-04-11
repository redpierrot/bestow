package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createTestFile(parentDir, content string, t *testing.T) string {
	testFile, err := os.CreateTemp(parentDir, "bestow_test_file_")
	if err != nil {
		t.Fatalf("failed to create the test file at %s; %v", parentDir, err)
	}
	_, err = testFile.WriteString(content)
	if err != nil {
		t.Fatalf("failed to write to the test file: %s; %v", testFile.Name(), err)
	}
	err = testFile.Close()
	if err != nil {
		t.Errorf("failed to close the test file: %v", err)
	}
	return testFile.Name()
}

func removeTestFile(path string) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	return nil
}

func createTempDirStructure() {
	tempDir, err := os.MkdirTemp(".", "bestow_test_*")
	if err != nil {
		fmt.Println("failed to create the temp directory for testing")
		os.Exit(1)
	}
	absPath, err := filepath.Abs(tempDir)
	if err != nil {
		fmt.Println("failed to retrieve the absolute path of the test directory")
		os.Exit(1)
	}
	testDir = &absPath
}

func createDirectory(parent, name string, t *testing.T) string {
	path := filepath.Join(parent, name)
	if err := os.Mkdir(path, 0744); err != nil {
		t.Fatalf("failed to create the directory; %v", err)
	}
	return path
}

func createTestDirectory(testName string, t *testing.T) testDirectory {
	dirPattern := fmt.Sprintf("%s_", testName)
	dirPath, err := os.MkdirTemp(*testDir, dirPattern)
	if err != nil {
		t.Fatalf("failed to create test subdirectory: %v", err)
	}
	srcDir := filepath.Join(dirPath, "src")
	destDir := filepath.Join(dirPath, "dest")
	if err := os.Mkdir(srcDir, 0700); err != nil {
		t.Fatalf("failed to create the src directory: %s; %v", srcDir, err)
	}
	if err := os.Mkdir(destDir, 0700); err != nil {
		t.Fatalf("failed to create the dest directory: %s; %v", destDir, err)
	}
	return testDirectory{
		parent: dirPath,
		src:    srcDir,
		dest:   destDir,
	}
}

func createTestPackage(packageName, src string, t *testing.T) []string {
	testFiles := []string{}
	packagePath := filepath.Join(src, packageName)
	if err := CreateDir(packagePath); err != nil {
		t.Fatalf("failed to create the test package; %v", err)
	}
	configDirPath := filepath.Join(packagePath, ".config")
	if err := CreateDir(configDirPath); err != nil {
		t.Fatalf("failed to create the config directory; %v", err)
	}
	content := fmt.Sprintf("sample content: %s", packageName)
	testFile := createTestFile(packagePath, content, t)
	testFiles = append(testFiles, testFile)
	configFile := createTestFile(configDirPath, content, t)
	testFiles = append(testFiles, configFile)
	return testFiles
}

func setPerm(path string, isPermitted bool, t *testing.T) {
	var mode os.FileMode
	if isPermitted {
		mode = 0644
	} else {
		mode = 0000
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("failed to set the permissions for the path: %s; %v", path, err)
	}
}
