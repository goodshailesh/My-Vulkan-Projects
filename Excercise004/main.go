package main

import (
	"fmt"
	"runtime"

	"github.com/vulkan-go/glfw/v3.3/glfw"
)

// Variable declaration
var win *glfw.Window

func main() {
	initWindow()
	//initVulkan()
	mainLoop()
	cleanup()
}

func initWindow() {

	runtime.LockOSThread()
	err := glfw.Init()
	if err != nil {
		panic(fmt.Errorf("Filed to initialize GLFW with error: %v", err))
	}
	// https://vulkan-tutorial.com/Drawing_a_triangle/Setup/Base_code
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI) // Because GLFW was originally designed to create an OpenGL context, we need to tell it to not create an OpenGL context with a subsequent cal
	// glfw.WindowHint(glfw.ContextVersionMajor, 4)
	// glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	//glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	//glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err = glfw.CreateWindow(800, 600, "Vulkan", nil, nil)
}

func mainLoop() {
	for !win.ShouldClose() {
		//win.SwapBuffers()
		glfw.PollEvents()
	}
}

func cleanup() {
	win.Destroy()
	glfw.Terminate()
}
