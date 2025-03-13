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
		PathName: strings.Join(paths, "/"),
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
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.PathName)
	_, err := os.Stat(pathNameWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

func (s *Store) Delete(key string) error {
	pathkey := s.PathNameTransFunc(key)

	defer func() {
		fmt.Printf("Removed the %+v From The Disk\n", pathkey.FileName)
	}()

	return os.RemoveAll(s.Root + "/" + pathkey.GetFirstRootPath())
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

// This will Open The File to Read the Contents Above
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	pathKey := s.PathNameTransFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", s.Root, pathKey.GetPathName())
	return os.Open(pathNameWithRoot)
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	fmt.Printf("Writing %+v To The Disk Of %+v On Server: %+v\n", key, r, s.Root)
	return s.writeStream(key, r)
}

// Writing the input reader io content to a Specific Pthaname and file name customized and copy them to the newly created file.
func (store *Store) writeStream(key string, r io.Reader) (int64, error) {
	//Gets the PathName
	pathKey := store.PathNameTransFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s", store.Root, pathKey.PathName)

	//Making A Directory
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return 0, err
	}

	//Gets the PathName With Proper styling
	pathAndFileName := pathKey.GetPathName()
	pathNameAndFileNameWithRoot := fmt.Sprintf("%s/%s", store.Root, pathAndFileName)
	//Creates a File
	file, err := os.Create(pathNameAndFileNameWithRoot)
	if err != nil {
		return 0, err
	}

	//Closes the file on function end using Defer
	defer file.Close()

	//Now Copies the File Content From the Reader. Also, Update is working well
	nb, err := io.Copy(file, r)
	if err != nil {
		return 0, err
	}

	fmt.Printf("%d bytes written to %s\n", nb, pathNameAndFileNameWithRoot)
	return nb, nil
}
