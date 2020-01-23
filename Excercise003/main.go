package main

import (
	"fmt"
	"runtime"

	"github.com/vulkan-go/glfw/v3.3/glfw"
)

func main() {
	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		panic(fmt.Errorf("Filed to initialize GLFW with error: %v", err))
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	//glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	win, err := glfw.CreateWindow(800, 600, "Vulkan", nil, nil)
	if err != nil {
		panic(fmt.Errorf("Failed to create Window with error : %v", err))
	}
	win.MakeContextCurrent()

	for !win.ShouldClose() {
		win.SwapBuffers()
		glfw.PollEvents()
	}
}
