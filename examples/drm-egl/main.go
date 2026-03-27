//go:build linux && cgo

package main

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/drmeglbackend"
	"github.com/AllenDang/cimgui-go/examples/common"
	"github.com/AllenDang/cimgui-go/imgui"
)

type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

var displayW int32
var displayH int32

var currentBackend backend.Backend[drmeglbackend.DRMEGLWindowFlags]

func init() {
	runtime.LockOSThread()
}

func startTouchInput() {
	f, err := os.OpenFile("/dev/input/event3", os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Error: unable open input device: %v\n", err)
		return
	}
	defer f.Close()

	var lastX, lastY float32

	evSize := int(unsafe.Sizeof(inputEvent{}))
	b := make([]byte, evSize)

	for {
		_, err := f.Read(b)
		if err != nil {
			continue
		}

		ev := (*inputEvent)(unsafe.Pointer(&b[0]))
		io := imgui.CurrentIO()
		if io == nil {
			continue
		}

		switch ev.Type {
		case 0x03: // EV_ABS
			switch ev.Code {
			case 0x35: // ABS_MT_POSITION_X
				lastX = float32(ev.Value) 
				io.AddMousePosEvent(lastX, lastY)
			case 0x36: // ABS_MT_POSITION_Y
				lastY = float32(ev.Value)
				io.AddMousePosEvent(lastX, lastY)
			}
		case 0x01: // EV_KEY
			if ev.Code == 0x14a { // BTN_TOUCH
				io.AddMouseButtonEvent(0, ev.Value == 1)
			}
		}
	}
}

func loop() {
	imgui.Begin("Touch Debug")
	imgui.Text(fmt.Sprintf("Application average %.3f ms/frame (%.1f FPS)", 1000.0/imgui.CurrentIO().Framerate(), imgui.CurrentIO().Framerate()))
	
	if imgui.Button("Test Button") {
		fmt.Println("button click")
	}
	
	imgui.End()
	common.Loop()
}

func main() {
	common.Initialize()
	drmbackend:=drmeglbackend.NewDRMEGLBackend()
	currentBackend, _ = backend.CreateBackend(drmbackend)
	currentBackend.SetAfterCreateContextHook(common.AfterCreateContext)
	currentBackend.SetBeforeDestroyContextHook(common.BeforeDestroyContext)
	currentBackend.SetBgColor(imgui.NewVec4(0.08, 0.10, 0.12, 1.0))
	currentBackend.CreateWindow("Cimgui-go DRM Touch", 0, 0)

	displayW, displayH = drmbackend.DisplaySize()
	fmt.Printf("real window size w %d h %d",displayW,displayH)
	go startTouchInput()

	currentBackend.Run(loop)
}