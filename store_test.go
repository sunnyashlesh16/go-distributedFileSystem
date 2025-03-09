package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	t.Skip("Skipping This")
	key := "GettingAJobSoon"
	pathKey := CASPathTransformFunc(key)
	expectedFileName := "22843397efdcd09b6252701c46940"
	expectedPathName := "sat/22843/397ef/dcd09/b6252/701c4/6940"
	if pathKey.PathName != expectedPathName {
		fmt.Errorf("PathTransformFunc returned %s instead of %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.FileName != expectedFileName {
		fmt.Errorf("PathTransformFunc returned %s instead of %s", pathKey.FileName, expectedFileName)
	}
}

func TestDelete(t *testing.T) {
	t.Skip("Skipping This")
	storeopts := StoreOpts{
		PathNameTransFunc: CASPathTransformFunc,
	}
	s := NewStore(storeopts)

	err := s.Delete("SampleCheck")
	if err != nil {
		fmt.Errorf("Delete returned an unexpected error: %s", err)
	}
}

func TestHas(t *testing.T) {
	//t.Skip("Skipping This")
	storeopts := StoreOpts{
		PathNameTransFunc: CASPathTransformFunc,
	}
	s := NewStore(storeopts)

	if ok := s.HasFile("SampleCheck"); !ok {
		t.Errorf("Respective File Is Not Found")
	}
}

func TestStore(t *testing.T) {
	t.Skip("Skipping This")
	storeopts := StoreOpts{
		PathNameTransFunc: CASPathTransformFunc,
	}
	s := NewStore(storeopts)

	defer s.Clear()

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("GettingAJobSoonInNext-%dDays", i)
		//I Don't Know But Some Way the Value or the contents are Being updated!
		data := []byte("Sunny, Keep Going You Will Get A Job Soon!")

		if err := s.writeStream(key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		// NB Will Be Ignored We Will Deal this later
		buf, _, err := s.Read(key)
		if err != nil {
			t.Error(err)
		}

		text, _ := io.ReadAll(buf)

		if string(text) != string(data) {
			fmt.Errorf("Store returned %s instead of %s", string(text), string(data))
		}

		s.Delete(key)
	}
}
