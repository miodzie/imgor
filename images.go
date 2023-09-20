package main

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"io"
	"math/rand"
	"os"
	"path"
)

type Image struct {
	Name     string
	Size     int64
	Fullpath string // todo it's probably better not to expose this since it can leak file path stuff D:
}

func (i *Image) File() (*os.File, error) {
	return os.Open(path.Join(i.Fullpath))
}

func (i *Image) Url() string {
	return domain + "/" + i.Name
}

type ImageStorage interface {
	Images() ([]*Image, error)
	Upload(file *os.File) (*Image, error) // maybe consider io.ReadCloser instead of os.File?
}

type FileStorage struct {
	BaseDir string
}

func (f *FileStorage) Upload(file *os.File) (*Image, error) {
	//TODO implement me
	panic("implement me")
}

func uploadFile(file io.ReadCloser, name string) (filename string, err error) {
	var b []byte
	b, _ = io.ReadAll(file)
	defer file.Close()
	file = io.NopCloser(bytes.NewBuffer(b))

	img, format, err := image.Decode(file)
	if err != nil {
		return
	}

	filename = name + "." + format

	f, err := os.Create(filesDir + "/" + filename)
	if err != nil {
		return
	}
	defer f.Close()
	switch format {
	case "gif":
		var g *gif.GIF
		g, err = gif.DecodeAll(bytes.NewBuffer(b))
		if err != nil {
			return
		}
		err = gif.EncodeAll(f, g)
		if err != nil {
			return
		}
	case "jpeg":
		err = jpeg.Encode(f, img, nil)
		if err != nil {
			return
		}
	case "png":
		err = pngEncoder.Encode(f, img)
		if err != nil {
			return
		}
	}
	return
}

func (f *FileStorage) Images() (images []*Image, err error) {
	var files []os.DirEntry
	files, err = os.ReadDir(f.BaseDir)
	if err != nil {
		return
	}

	for _, file := range files {
		var info os.FileInfo
		info, err = file.Info()
		if err != nil {
			return
		}
		images = append(images, &Image{
			Name:     info.Name(),
			Size:     info.Size(),
			Fullpath: path.Join(f.BaseDir, file.Name()),
		})
	}

	return
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randId() string {
	n := 16
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
