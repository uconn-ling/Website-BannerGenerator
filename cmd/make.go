// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/anthonynsimon/bild/imgio"
	"github.com/anthonynsimon/bild/transform"
	"github.com/fogleman/gg"
	"github.com/spf13/cobra"
)

// makeCmd represents the make command
var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "run make to build banners",
	Long:  `put pictures in full size in a filder named "inputPictures"`,
	Run:   mainBuild,
}

const (
	setWidth  = 950
	setHeight = 320
	// W width
	W = setWidth * 4
	// H height
	H             = setHeight * 4
	pictureCount  = 8
	maxRotate     = 7
	border        = 5 * 4
	rows          = 2
	percentBorder = 0.05
)

var (
	tempDir string
	maxWH   int
)

func init() {
	rootCmd.AddCommand(makeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// makeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// makeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func shuffle(files []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(files), func(i, j int) { files[i], files[j] = files[j], files[i] })
	return files
}

func buildArrays(pictureCount int, newFiles []string) [][]string {
	newFiles = shuffle(newFiles)
	// fmt.Println(len(files), len(files)%pictureCount)
	addFiles := len(newFiles) % pictureCount
	normalFiles := len(newFiles) / pictureCount
	var twoD [][]string
	if addFiles > 0 {

		var temp []string
		count := 0

		for count < pictureCount {
			if count < addFiles {
				var x string
				x, newFiles = newFiles[0], newFiles[1:]

				temp = append(temp, x)
			} else {
				temp = append(temp, newFiles[(count-addFiles)])
			}
			count++
		}
		newFiles = shuffle(newFiles)
		twoD = append(twoD, temp)
	}

	count := 0

	for count < normalFiles {
		var temp []string
		subcount := 0
		for subcount < pictureCount {
			var x string
			x, newFiles = newFiles[0], newFiles[1:]

			temp = append(temp, x)
			subcount++
		}
		twoD = append(twoD, temp)
		count++
	}
	return twoD
}

func preparePicture(singleFileName string) {
	im, err := gg.LoadImage("inputPictures/" + singleFileName)
	if err != nil {
		panic(err)
	}

	iw, ih := im.Bounds().Dx(), im.Bounds().Dy()
	var newWith int
	var newHeight int
	if iw > ih {
		newWith = maxWH
		newHeight = (newWith * ih) / iw
	} else {
		newHeight = maxWH
		newWith = (newHeight * iw) / ih
	}

	dc := gg.NewContext(newWith+(border*2)+(border/2), newHeight+(border*2)+(border/2))
	dc.DrawRectangle((border / 2), (border / 2), float64(newWith+(border*2)+(border/2)), float64(newHeight+(border*2)+(border/2)))
	dc.SetRGBA(0, 0, 0, 0.2)
	dc.Fill()
	dc.DrawRectangle(0, 0, float64(newWith+(border*2)), float64(newHeight+(border*2)))
	dc.SetHexColor("FFFFFF")
	dc.Fill()

	dc.DrawLine(float64(0), float64(1), float64(0), float64(newHeight+(border*2)))
	dc.DrawLine(float64(newWith+(border*2)), float64(1), float64(newWith+(border*2)), float64(newHeight+(border*2)))
	dc.DrawLine(float64(1), float64(0), float64(newWith+(border*2)), float64(0))
	dc.DrawLine(float64(1), float64(newHeight+(border*2)), float64(newWith+(border*2)), float64(newHeight+(border*2)))
	dc.SetLineWidth(0.5)
	dc.SetRGB(200, 200, 200)
	dc.Stroke()
	resized := transform.Resize(im, newWith, newHeight, transform.Linear)

	dc.DrawImage(resized, border, border)

	dc.SavePNG(path.Join(tempDir, singleFileName+".png"))
}

func mainBuild(cmd *cobra.Command, args []string) {
	if W > H {
		divA := ((pictureCount / rows) + 0.5)
		maxWH = int(W / divA)
	} else {
		divA := (rows + 0.5)
		maxWH = int(H / divA)
	}
	dir, err := ioutil.TempDir("", "bannerGen")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tempDir = dir

	files, err := ioutil.ReadDir(path.Join("./inputPictures/"))
	if err != nil {
		log.Fatal(err)
	}
	var newFiles []string
	for i := 0; i < len(files); i++ {
		if strings.Compare(files[i].Name(), "dummy") != 0 && !files[i].IsDir() {
			newFiles = append(newFiles, files[i].Name())
		}
	}

	twoD := buildArrays(pictureCount, newFiles)
	// fmt.Println(twoD)

	os.Mkdir(path.Join("./output/"), os.FileMode(0722))
	os.Mkdir(path.Join("./output/big/"), os.FileMode(0722))
	for j, singleFileArray := range twoD {
		for _, singleFileName := range singleFileArray {
			preparePicture(singleFileName)
		}
		complete := gg.NewContext(W, H)
		complete.SetHexColor("FFFFFF")
		complete.Clear()
		colCount := 0
		rowCount := 0
		for _, singleFileName := range singleFileArray {
			if colCount >= (pictureCount / rows) {
				rowCount++
				colCount = 0
			}
			sim, err := gg.LoadImage(path.Join(tempDir, singleFileName+".png"))
			if err != nil {
				panic(err)
			}
			rotate := rand.Intn((maxRotate))
			if colCount%2 > 0 {
				// fmt.Println("i", i, i%2)
				rotate = rotate * (-1)
			}
			if rowCount%2 > 0 {
				// fmt.Println("rowCount", rowCount, rowCount%2)
				rotate = rotate * (-1)
			}
			// fmt.Println("rotate", rotate)
			rotated := transform.Rotate(sim, float64(rotate), &transform.RotationOptions{ResizeBounds: true})
			xPart := (W / (pictureCount / rows)) / 2
			yPart := (H / rows) / 2
			colAddRemov := 0.0
			rowAddRemov := 0.0

			if rowCount == 0 {
				rowAddRemov = float64(maxWH) * percentBorder
			} else if rowCount == (rows - 1) {
				rowAddRemov = -float64(maxWH) * percentBorder
			} else if rowCount < (rows / 2) {
				rowAddRemov = (float64(maxWH) * percentBorder) / 2
			} else {
				rowAddRemov = -(float64(maxWH) * percentBorder) / 2
			}
			if colCount == 0 {
				colAddRemov = float64(maxWH) * percentBorder
			} else if colCount == ((pictureCount / rows) - 1) {
				colAddRemov = -float64(maxWH) * percentBorder
			} else if colCount < ((pictureCount / rows) / 2) {
				colAddRemov = (float64(maxWH) * percentBorder) / 2
			} else {
				colAddRemov = -(float64(maxWH) * percentBorder) / 2
			}
			placeX := xPart + (colCount)*(W/(pictureCount/rows)) + int(colAddRemov)
			placeY := yPart + (rowCount)*(H/rows) + int(rowAddRemov)
			// fmt.Println(placeX, placeY)
			complete.DrawImageAnchored(rotated, placeX, placeY, 0.5, 0.5)
			colCount++
		}
		newPathBig := path.Join("./output/big/", "banner-"+strconv.Itoa(j+1)+".png")
		newPath := path.Join("./output/", "banner-"+strconv.Itoa(j+1)+".jpg")
		fmt.Println(newPathBig)
		complete.SavePNG(newPathBig)
		sim, err := gg.LoadImage(newPathBig)
		if err != nil {
			panic(err)
		}
		resized := transform.Resize(sim, setWidth, setHeight, transform.Linear)
		if err := imgio.Save(newPath, resized, imgio.JPEGEncoder(95)); err != nil {
			panic(err)
		}
	}
}
