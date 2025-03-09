package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const defaultRootPathName = "sat"

// Transformation Of Storing Pattern based on the Key Provided On the Disk
func CASPathTransformFunc(key string) PathKey {
	hash := md5.Sum([]byte(key))
	Hash := hex.EncodeToString(hash[:])

	blockSize := 5
	SliceLen := len(Hash) / blockSize
	paths := make([]string, SliceLen)

	for i := 0; i < SliceLen; i++ {
		from, end := i*blockSize, (i*blockSize)+blockSize
		paths[i] = Hash[from:end]
	}

	return PathKey{
		PathName: defaultRootPathName + "/" + strings.Join(paths, "/"),
		FileName: Hash,
	}
}

type PathNameTransFunc func(string) PathKey

func (p PathKey) GetPathName() string {
	return fmt.Sprintf("%s%s", p.PathName, p.FileName)
}

func (p PathKey) GetFirstRootPath() string {
	paths := strings.Split(p.PathName, "/")
	return paths[1]
}

func DefaultPathTransFunc(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

type PathKey struct {
	PathName string
	FileName string
}

// To Set The Storage Options
type StoreOpts struct {
	Root              string
	PathNameTransFunc PathNameTransFunc
}

// To Set The Store
type Store struct {
	StoreOpts
}

// Created A New Store or Creation Of New Store Based On Store Options
func NewStore(opts StoreOpts) *Store {
	//Setting The default Trans if the Opts is nil!
	if opts.PathNameTransFunc == nil {
		opts.PathNameTransFunc = DefaultPathTransFunc
	}
	if opts.Root == "" {
		opts.Root = defaultRootPathName
	}
	return &Store{opts}
}

func (s *Store) HasFile(key string) bool {
	pathKey := s.PathNameTransFunc(key)
	_, err := os.Stat(pathKey.PathName)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}
func (s *Store) Delete(key string) error {
	pathkey := s.PathNameTransFunc(key)

	defer func() {
		fmt.Printf("Removed the %+v From The Disk", pathkey.FileName)
	}()

	return os.RemoveAll(defaultRootPathName + "/" + pathkey.GetFirstRootPath())
}

// To read the Contents Of The File based on the key value provided, When Don't Know About the type just use any!
func (s *Store) Read(key string) (io.Reader, any, error) {
	file, err := s.readStream(key)
	if err != nil {
		return nil, 0, err
	}

	defer file.Close()

	buf := new(bytes.Buffer)
	//Ignore the other values other than the error! We can USe (_)
	nb, err := io.Copy(buf, file)

	return buf, nb, err
}

// This will Open The File to Read the COntents Above
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathNameTransFunc(key)
	return os.Open(pathKey.GetPathName())
}

func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

// Writing the input reader io content to a Specific Pthaname and file name customized and copy them to the newly created file.
func (store *Store) writeStream(key string, r io.Reader) error {
	//Gets the PathName
	pathKey := store.PathNameTransFunc(key)

	//Making A Directory
	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	//Gets the PathName With Proper styling
	pathAndFileName := pathKey.GetPathName()

	//Creates a File
	file, err := os.Create(pathAndFileName)
	if err != nil {
		return err
	}

	//Closes the file on function end using Defer
	defer file.Close()

	//Now Copies the File Content From the Reader. Also, Update is working well
	nb, err := io.Copy(file, r)
	if err != nil {
		return err
	}

	fmt.Printf("%d bytes written to %s\n", nb, pathAndFileName)
	return nil
}
