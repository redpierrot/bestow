package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ThisaruGuruge/bestow/internal/log"
)

var testDir *string

type testDirectory struct {
	parent string
	src    string
	dest   string
}

func TestMain(m *testing.M) {
	fmt.Println("Running File Tests")
	log.SetLogger(log.NewCharmLogger())

	createTempDirStructure()
	log.Info("Created the temp directory", "test_directory_path", *testDir)

	code := m.Run()

	log.Info("Finished File Tests. Cleaning up")
	os.RemoveAll(*testDir)

	os.Exit(code)
}

func TestListAllDirectories(t *testing.T) {
	testDirs := createTestDirectory("test_all_directories", t)
	dir := testDirs.src
	fileList, err := ListAllDirectories(dir)
	if err != nil {
		t.Fatalf("failed to read directory %v", err)
	}
	if len(fileList) != 0 {
		t.Fatalf("temp directory [%s] is not empty", dir)
	}
	for i := range 5 {
		subDir := fmt.Sprintf("dir_%d", i)
		subDirPath := filepath.Join(dir, subDir)
		if err := os.Mkdir(subDirPath, 0644); err != nil {
			t.Fatalf("failed to create temp test directory: %s; %v", subDirPath, err)
		}
	}
	fileList, err = ListAllDirectories(dir)
	if err != nil {
		t.Fatalf("failed to read the test directory: %v", err)
	}
	if len(fileList) != 5 {
		t.Fatalf("directory list mismatch; expected: %d actual: %d", 5, len(fileList))
	}
}

func TestCreateDir(t *testing.T) {
	dir := createTestDirectory("test_create_dir", t).src
	path := filepath.Join(dir, "package1")
	err := CreateDir(path)
	if err != nil {
		t.Fatalf("failed to create directory %s", path)
	}
	fileList, err := ListAllDirectories(dir)
	if err != nil {
		t.Fatalf("failed to read the directory %s", dir)
	}
	if !slices.Contains(fileList, "package1") {
		t.Fatalf("directory creation failed %s", path)
	}
}

func TestIsDir(t *testing.T) {
	dir := createTestDirectory("test_is_dir", t).parent
	isDir, err := IsDir(dir)
	if err != nil {
		t.Fatalf("failed to read the directory; %v", err)
	}
	if !isDir {
		t.Fatalf("isDir check failed for the directory: %s", dir)
	}
}

func TestCreateFile(t *testing.T) {
	dir := createTestDirectory("test_create_file", t).src
	testFile := "test_file.txt"
	testFileContent := "sample file for testing"
	if err := CreateFile(testFile, dir, testFileContent); err != nil {
		t.Fatalf("failed to create the test file: %v", err)
	}
	lines, err := ReadLines(filepath.Join(dir, testFile))
	if err != nil {
		t.Fatalf("failed to read the file: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("file content mismatch; expected: %s actual: [%s]", testFileContent, strings.Join(lines, ", "))
	}
	if lines[0] != testFileContent {
		t.Fatalf("file content mismatch; expected: %s actual: %s", testFileContent, lines[0])
	}
}

func TestGetPathSegments(t *testing.T) {
	dir := createTestDirectory("test_path_segments", t).parent
	log.Info("Directory Path", "path", dir)
	segments := GetPathSegments(dir)
	log.Info("recieved the path segments", "segments", segments)
}

func TestWriteFile(t *testing.T) {
	testDir := createTestDirectory("test_write_file", t)
	createDirectory(testDir.src, "sample", t)
}

func TestLink(t *testing.T) {
	testDirs := createTestDirectory("test_link", t)

	fileName := "link_test.txt"
	fileContent := "sample file for testing"
	destFilePath := filepath.Join(testDirs.dest, fileName)
	srcFilePath := filepath.Join(testDirs.src, fileName)
	if err := os.WriteFile(srcFilePath, []byte(fileContent), 0644); err != nil {
		t.Fatalf("failed to write to test file: %s; %v", srcFilePath, err)
	}
	if err := Link(srcFilePath, destFilePath); err != nil {
		t.Fatalf("failed to link the file: %v", err)
	}
	srcStat, err := os.Stat(srcFilePath)
	if err != nil {
		t.Fatalf("failed to read the source file: %s; %v", srcFilePath, err)
	}
	destStat, err := os.Stat(destFilePath)
	if err != nil {
		t.Fatalf("failed to read the destination file: %s; %v", destFilePath, err)
	}
	if !os.SameFile(srcStat, destStat) {
		t.Fatalf("link not created properly; files are different: source: %s destination: %s", srcFilePath, destFilePath)
	}
}

func TestReadLines(t *testing.T) {
	dir := createTestDirectory("test_read_lines", t).parent
	content := `
	this is a multi-line text
	to test reading lines
	from the test files
	`

	fileName := createTestFile(dir, content, t)
	lines, err := ReadLines(fileName)
	if err != nil {
		t.Fatalf("failed to read lines %v", err)
	}
	expectedLines := strings.Split(content, "\n")

	if len(lines) != len(expectedLines) {
		t.Fatalf("line count mismatch when reading file; expected: %d actual: %d", len(lines), len(expectedLines))
	}

	for i, line := range lines {
		if line != expectedLines[i] {
			t.Fatalf("content mismatch; expected: '%s', actual: '%s'", line, expectedLines[i])
		}
	}
	if err := removeTestFile(fileName); err != nil {
		t.Errorf("failed to remove the test file: %s; %v", fileName, err)
	}
}

func TestBackup(t *testing.T) {
	dir := createTestDirectory("test_backup", t).dest
	testFile := createTestFile(dir, "test file content", t)
	if err := Backup(testFile); err != nil {
		t.Fatalf("failed to backup the file: %s; %v", testFile, err)
	}
	exists, err := Exists(testFile)
	if err != nil {
		t.Fatalf("failed to check the test file: %s; %v", testFile, err)
	}
	if exists {
		t.Errorf("test file still exists, expected to backup: %s", testFile)
	}

	backupFile := testFile + ".bestow.bak"
	exists, err = Exists(backupFile)
	if err != nil {
		t.Fatalf("failed to check the backup file: %s; %v", backupFile, err)
	}
	if !exists {
		t.Fatalf("backup file does not exist: %s", backupFile)
	}
}

func TestCopy(t *testing.T) {
	testDir := createTestDirectory("test_copy", t)
	content := "test file content for copy function test"
	testFile := createTestFile(testDir.src, content, t)
	destFile := filepath.Join(testDir.dest, "copy_destination")
	if err := Copy(testFile, destFile); err != nil {
		t.Fatalf("failed to copy the file: %v", err)
	}

	destFileBytes, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("failed to read from the copied file: %v", err)
	}
	destFileConent := string(destFileBytes)
	if content != destFileConent {
		t.Fatalf("source and destination content mismatch; expected %s actual: %s", content, destFileConent)
	}
}

func TestRemove(t *testing.T) {
	testDir := createTestDirectory("test_remove", t)
	dir := testDir.src

	regular := createTestFile(dir, "remove test content", t)
	noPermParent := createDirectory(dir, "no_perm", t)
	noPerm := createTestFile(noPermParent, "no perm file", t)
	setPerm(noPermParent, false, t)
	defer setPerm(noPermParent, true, t)
	nonExistentFileError := &FileError{
		Message: "failed to remove the existing symlink/file",
		Path:    "non_existent.file",
		Cause:   errors.New("remove non_existent.file: no such file or directory"),
	}
	noPermError := &FileError{
		Message: "failed to remove the existing symlink/file",
		Path:    noPerm,
		Cause:   errors.New(fmt.Sprintf("remove %s: permission denied", noPerm)),
	}

	tests := map[string]struct {
		src string
		err error
	}{
		"regular": {
			src: regular,
			err: nil,
		},
		"non_existent": {
			src: "non_existent.file",
			err: nonExistentFileError,
		},
		"no_perm": {
			src: noPerm,
			err: noPermError,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Logf("running: %s", name)
			if err := Remove(tc.src); err != nil {
				t.Logf("error returned: file: %s", tc.src)
				t.Logf("%v", err)
				if tc.err != nil {
					if err.Error() != tc.err.Error() {
						t.Logf("%v", tc.err)
						t.Fatalf("error type mismatch: expected: %v actual: %v", tc.err, err)
					}
				} else {
					t.Fatalf("failed to remove the test file: %s; %v", tc.src, err)
				}
			} else {
				t.Logf("no error occurred: file: %s", tc.src)
				if tc.err != nil {
					t.Fatalf("expected error: %v", tc.err)
				}
				exists, err := Exists(tc.src)
				if err != nil {
					t.Errorf("failed to check the test file: %s; %v", tc.src, err)
				}
				if exists {
					t.Fatalf("failed to remove the test file: %s", tc.src)
				}
			}
		})
	}
}

func TestGetExistingFileType(t *testing.T) {
	testDir := createTestDirectory("get_existing_type", t)

	// Create files
	regularFile := createTestFile(testDir.dest, "test file", t)
	symlinkOriginal := createTestFile(testDir.src, "sample link source file", t)
	nonLinkedSrc := createTestFile(testDir.src, "sample non link source file", t)
	destSymlinkPath := filepath.Join(testDir.dest, filepath.Base(symlinkOriginal))
	if err := Link(symlinkOriginal, destSymlinkPath); err != nil {
		t.Fatalf("failed to create the symlink: %s; %v", destSymlinkPath, err)
	}
	regularFileNoPerms := createTestFile(testDir.dest, "test file no perm", t)
	nonExistentPath := filepath.Join(testDir.dest, "non_existent")
	_, cause := os.Lstat(nonExistentPath)
	nonExistingPathError := FileError{
		Message: "failed to read the path",
		Path:    nonExistentPath,
		Cause:   cause,
	}

	tests := map[string]struct {
		src      string
		dest     string
		expected ExistingType
		err      error
	}{
		"regular_file": {
			src:      filepath.Join(testDir.src, filepath.Base(regularFile)),
			dest:     regularFile,
			expected: ExistingRegularFile,
			err:      nil,
		},
		"managed_symlink": {
			src:      symlinkOriginal,
			dest:     destSymlinkPath,
			expected: ExistingManagedSymlink,
			err:      nil,
		},
		"existing_dir": {
			src:      testDir.src,
			dest:     testDir.dest,
			expected: ExistingDir,
			err:      nil,
		},
		"foreign_symlink": {
			src:      nonLinkedSrc,
			dest:     destSymlinkPath,
			expected: ExistingForeignSymlink,
			err:      nil,
		},
		"regular_file_no_perm": {
			src:      nonLinkedSrc,
			dest:     regularFileNoPerms,
			expected: ExistingRegularFile,
			err:      nil,
		},
		"non_existing_dest": {
			src:      nonLinkedSrc,
			dest:     filepath.Join(testDir.dest, "non_existent"),
			expected: ExistingRegularFile,
			err:      &nonExistingPathError,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := GetExistingFileType(tc.src, tc.dest)
			if err != nil {
				if tc.err == nil {
					t.Fatalf("failed to retrieve the existing file type; %v", err)
				}
				if err.Error() != tc.err.Error() {
					t.Fatalf("error type mismatch: expected %v actual %v", tc.err, err)
				}
			} else {
				if tc.err != nil {
					t.Fatalf("expected error: expected: %v actual: nil", tc.err)
				}
				if actual != tc.expected {
					t.Fatalf("expected value mismatch: expected: %s actual: %s", tc.expected, actual)
				}
				log.Info("[TestGetExistingFileType]: test successful", "case", name)
			}
		})
	}
}

func TestListAllFilesInDir(t *testing.T) {
	testDir := createTestDirectory("test_list_all_files_in_dir", t)
	testFiles := []string{}
	for i := range 5 {
		packageName := fmt.Sprintf("package_%d", i)
		packageFiles := createTestPackage(packageName, testDir.src, t)
		testFiles = append(testFiles, packageFiles...)
	}

	testFileNames := []string{}
	for _, fileName := range testFiles {
		// Get relative path from the parent (simulate source)
		fileRelPath := strings.Replace(fileName, testDir.parent, "", 1)
		// Remove leading separator
		fileRelPath = strings.TrimLeft(fileRelPath, string(filepath.Separator))
		testFileNames = append(testFileNames, fileRelPath)
	}

	nonExistentDirName := "non_existent"
	nonExistentPath := filepath.Join(testDir.parent, nonExistentDirName)
	_, notFoundErr := os.Stat(nonExistentPath)

	tests := map[string]struct {
		parent string
		dir    string
		files  []string
		err    error
	}{
		"list_dir": {
			parent: testDir.parent,
			dir:    "src",
			files:  testFileNames,
			err:    nil,
		},
		"invalid_dir": {
			parent: testDir.parent,
			dir:    "non_existent",
			files:  []string{},
			err: &FileError{
				Message: "provided directory path not found",
				Path:    nonExistentPath,
				Cause:   notFoundErr,
			},
		},
		"empty_dir": {
			parent: "",
			dir:    "",
			files:  []string{},
			err: &FileError{
				Message: "path name is empty",
				Path:    "",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fileList, err := ListAllFilesInDir(tc.parent, tc.dir)
			fileNames := []string{}
			for _, fileName := range fileList {
				fileNames = append(fileNames, filepath.Base(fileName))
			}
			if err != nil {
				if tc.err == nil {
					t.Fatalf("failed to read the files from the directory: %s; %v", testDir.parent, err)
				}
				if tc.err.Error() != err.Error() {
					t.Fatalf("error mismatch; expected: %v actual: %v", tc.err, err)
				}
			} else {
				if tc.err != nil {
					t.Fatalf("expected error %v", tc.err)
				}
				if len(fileList) != len(testFiles) {
					t.Fatalf("file count mismatch: expected %d, actual %d", len(testFiles), len(fileList))
				}
				for _, item := range tc.files {
					if !slices.Contains(fileList, item) {
						t.Fatalf("expected file list not found; expected: %s actual: %s", strings.Join(tc.files, ", "), strings.Join(fileList, ", "))
					}
				}
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	testDir := createTestDirectory("test_list_files", t)
	testFiles := []string{}
	for _ = range 5 {
		fileName := createTestFile(testDir.src, "sample content", t)
		testFiles = append(testFiles, filepath.Base(fileName))
	}

	nonDir := filepath.Join(testDir.src, testFiles[0])

	tests := map[string]struct {
		path string
		err  error
	}{
		"list_files": {
			path: testDir.src,
			err:  nil,
		},
		"non_dir": {
			path: nonDir,
			err: &FileError{
				Message: "provided path is not a directory",
				Path:    nonDir,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fileList, err := ListFiles(tc.path)
			if err != nil {
				if tc.err == nil {
					t.Fatalf("failed to list the files; %v", err)
				}
				if tc.err.Error() != err.Error() {
					t.Fatalf("error mismatch; expected: %v actual: %v", tc.err, err)
				}
			} else {
				if tc.err != nil {
					t.Fatalf("expected error: %v", tc.err)
				}
				if len(fileList) != len(testFiles) {
					t.Fatalf("file list count mismatch: expected: %d actual: %d", len(testFiles), len(fileList))
				}
				for _, item := range fileList {
					if !slices.Contains(testFiles, item) {
						t.Log("expected: ", strings.Join(testFiles, ", "))
						t.Log("actual: ", strings.Join(fileList, ", "))
						t.Fatalf("expected file not found: %s", item)
					}
				}
			}
		})
	}
}
