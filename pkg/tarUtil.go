package tarUtil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreateTar(source, tarName string, filesToIgnore []string) error {
	tarFile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return writeToTar(source, tarWriter, append(filesToIgnore, tarName))
}

func UnTar(source, destination string) error {
	tarFile, err := os.Open(source)
	if err != nil {
		return err
	}

	gzipReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header == nil {
			continue
		}
		target := filepath.Join(destination, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			fmt.Printf("Creating dir: %s\n", target)
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			if err := createParentDir(target); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err = io.Copy(file, tarFile); err != nil {
				return err
			}
			file.Close()
		}
	}
}

func createParentDir(file string) error {
	dir := filepath.Dir(file)
	_, err := os.Stat(dir)
	if err != nil {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

func matchesAny(source string, ignore []string) bool {
	for _, filePatternIgnore := range ignore {
		matched, err := filepath.Match(filePatternIgnore, strings.TrimPrefix(source, "./"))
		if err == nil && matched {
			return matched
		}
	}
	return false
}

func writeToTar(source string, writer *tar.Writer, ignore []string) error {
	if matchesAny(source, ignore) {
		println("Ignoring " + source)
		return nil
	}
	stat, err := os.Stat(source)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, stat.Name())
	if err != nil {
		return err
	}
	header.Name = source
	err = writer.WriteHeader(header)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		dirEntries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, dirEntry := range dirEntries {
			err := writeToTar(fmt.Sprintf("%s/%s", source, dirEntry.Name()), writer, ignore)
			if err != nil {
				return err
			}
		}
		return nil
	}
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	_, err = io.Copy(writer, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

func AddToTar(tarContext []byte, filepath, filename string) error {
	stat, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(bytes.NewBuffer(tarContext))
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	header, err := tar.FileInfoHeader(stat, stat.Name())
	header.Name = filename
	if err != nil {
		return err
	}
	err = tarWriter.WriteHeader(header)
	if err != nil {
		return err
	}

	additionalFile, err := os.Open(filepath)
	defer additionalFile.Close()

	if err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, additionalFile)
	return err
}

func ContainFile(tarFileContent []byte, filename string) bool {
	gzipReader, err := gzip.NewReader(bytes.NewBuffer(tarFileContent))
	if err != nil {
		return false
	}
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}
		if header.Name == filename {
			return true
		}
	}
	return false
}
