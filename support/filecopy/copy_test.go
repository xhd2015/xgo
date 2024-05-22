package filecopy

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func testCopyReplace(prepare func(rootDir string, srcDir string, dstDir string) error, check func(rootDir string, srcDir string, dstDir string) error) error {
	tmpDir, err := os.MkdirTemp("", "copy-with-link")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	err = os.MkdirAll(srcDir, 0755)
	if err != nil {
		return err
	}
	dstDir := filepath.Join(tmpDir, "dst")

	err = prepare(tmpDir, srcDir, dstDir)
	if err != nil {
		return err
	}
	err = CopyReplaceDir(srcDir, dstDir, false)
	if err != nil {
		return err
	}
	if check == nil {
		return nil
	}
	return check(tmpDir, srcDir, dstDir)
}

func TestCopyWithSymLinkFiles(t *testing.T) {
	// doc.txt
	// src/
	//    a.txt
	//    doc.txt -> ../doc.txt
	err := testCopyReplace(func(rootDir, srcDir, dstDir string) error {
		docTxt := filepath.Join(rootDir, "doc.txt")
		err := ioutil.WriteFile(docTxt, []byte("doc"), 0755)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("test"), 0755)
		if err != nil {
			return err
		}
		err = os.Symlink(docTxt, filepath.Join(srcDir, "doc.txt"))
		if err != nil {
			return err
		}
		return nil
	}, func(rootDir, srcDir, dstDir string) error {
		ok, err := checkIsSymLink(filepath.Join(dstDir, "doc.txt"))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("expect dst/doc.txt to be sym link, actually not")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func checkIsSymLink(file string) (bool, error) {
	finfo, err := os.Lstat(file)
	if err != nil {
		return false, err
	}
	if finfo.Mode()&fs.ModeSymlink != 0 {
		return true, nil
	}
	return false, nil
}
func TestCopyWithSymLinkDirs(t *testing.T) {
	// doc.txt
	// src/
	//    a.txt
	//    doc.txt -> ../doc.txt
	err := testCopyReplace(func(rootDir, srcDir, dstDir string) error {
		docDir := filepath.Join(rootDir, "doc")
		err := os.MkdirAll(docDir, 0755)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filepath.Join(docDir, "doc.txt"), []byte("doc"), 0755)
		if err != nil {
			return err
		}
		err = os.Symlink(docDir, filepath.Join(srcDir, "doc"))
		if err != nil {
			return err
		}
		return nil
	}, func(rootDir, srcDir, dstDir string) error {
		ok, err := checkIsSymLink(filepath.Join(dstDir, "doc"))
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("expect dst/doc to be sym link, actually not")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
