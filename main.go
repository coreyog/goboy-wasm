package main

import (
	"fmt"
	"image"
	"image/draw"
	"syscall/js"

	"github.com/coreyog/goboy-wasm/gradient"

	"golang.org/x/image/colornames"
)

const (
	width  = 160
	height = 144
)

var ctx js.Value
var requestAnimationFrame js.Value
var jsOnFrame js.Func
var console js.Value
var fps js.Value
var img *image.RGBA
var progress float64
var prevTS float64
var killSwitch chan struct{}
var closing bool

func main() {
	// prep state
	killSwitch = make(chan struct{})
	document := js.Global().Get("document")
	canvas := document.Call("getElementById", "target")

	fn := canvas.Get("getContext")
	if !fn.Truthy() {
		return
	}

	fps = document.Call("getElementById", "fps")
	console = js.Global().Get("console")
	ctx = canvas.Call("getContext", "2d")
	requestAnimationFrame = js.Global().Get("requestAnimationFrame")
	jsOnFrame = js.FuncOf(onFrame)

	// create image
	img = image.NewRGBA(image.Rect(0, 0, width, height))

	// 1px solid black border
	draw.Draw(img, img.Bounds(), image.NewUniform(colornames.Black), image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(1, 1, width-1, height-1), image.NewUniform(colornames.White), image.Point{}, draw.Src)

	// draw frame
	onFrame(js.Null(), []js.Value{js.ValueOf(0)})

	// register kill switch
	js.Global().Set("stopWASM", js.FuncOf(stopWASM))

	// wait for call to stopWASM
	<-killSwitch

	// fill the image with white and clear the canvas
	draw.Draw(img, image.Rect(0, 0, width, height), image.NewUniform(colornames.White), image.Point{}, draw.Src)
	drawImage(ctx, img)

	fps.Set("innerHTML", "---")
}

func onFrame(this js.Value, args []js.Value) interface{} {
	// guarenteed requestAnimationFrame or kill the app
	defer func() {
		if !closing {
			requestAnimationFrame.Invoke(jsOnFrame)
		} else {
			killSwitch <- struct{}{}
		}
	}()

	// determine timestamp and delta time
	ts := args[0].Float()      // in milliseconds since start
	dt := (ts - prevTS) / 1000 // in seconds
	prevTS = ts

	// update FPS in DOM
	text := fmt.Sprintf("fps: %0.0f\n", 1/dt)
	fps.Set("innerHTML", text)
	fmt.Println(text)

	// inset colored rectangle
	draw.Draw(img, image.Rect(10, 10, width-10, height-10), image.NewUniform(gradient.Keypoints.GetInterpolatedColorFor(progress)), image.Point{}, draw.Src)
	drawImage(ctx, img)

	// increment progress through the gradient
	progress += dt / 3
	if progress > 1 {
		progress -= 1
	}

	return js.Null()
}

func drawImage(ctx js.Value, img *image.RGBA) {
	data := make([]byte, width*height*4)

	// get pixel data
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := (y*width + x) * 4
			pixel := img.RGBAAt(x, y)
			data[index] = pixel.R
			data[index+1] = pixel.G
			data[index+2] = pixel.B
			data[index+3] = pixel.A
		}
	}

	// copy to JS
	Uint8Array := js.Global().Get("Uint8Array")
	jsData := Uint8Array.New(len(data))
	js.CopyBytesToJS(jsData, data)

	// clamp the data
	Uint8ClampedArray := js.Global().Get("Uint8ClampedArray")
	jsClampedData := Uint8ClampedArray.New(jsData) // view, don't use "Uint8ClampedArray.from(...)"

	// make it Image Data
	ImageData := js.Global().Get("ImageData")
	imgData := ImageData.New(jsClampedData, width)

	// put it on the canvas
	ctx.Call("putImageData", imgData, 0, 0)
}

func stopWASM(this js.Value, args []js.Value) interface{} {
	closing = true
	return js.Null()
}
