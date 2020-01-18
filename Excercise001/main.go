package main

import (
	"fmt"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	//-------------------------
	// Initialization Stage
	//-------------------------

	orPanic(glfw.Init())
	// GetVulkanGetInstanceProcAddress returns the function pointer used to find Vulkan core or
	// extension functions. The return value of this function can be passed to the Vulkan library.
	// Note that this function does not work the same way as the glfwGetInstanceProcAddress.
	// https://github.com/vulkan-go/glfw/blob/v3.3/v3.3/glfw/vulkan.go
	// SetGetInstanceProcAddr sets the GetInstanceProcAddr function pointer used to load Vulkan symbols.
	// This function must be called before vulkan.Init()
	// https://godoc.org/github.com/vulkan-go/vulkan#SetGetInstanceProcAddr
	//fmt.Println(glfw.GetVulkanGetInstanceProcAddress())
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	orPanic(vk.Init())

	// *** 1 List Layers available ***//

	xPrintInstanceLayerProperties()
	xPrintInstanceExtensionProperties()

	// *** 2 Instance Creation ***//

	instance := xCreateInstance()
	fmt.Println(instance)

}

func xCreateInstance() vk.Instance {
	var instance vk.Instance
	var layers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var extensions = []string{"VK_KHR_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PNext: nil,
		//Flags:                   nil,
		PApplicationInfo:        nil,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	orPanic(vk.CreateInstance(&instanceInfo, nil, &instance))
	return instance
}

func xPrintInstanceExtensionProperties() {
	var extensionCount uint32
	vk.EnumerateInstanceExtensionProperties("", &extensionCount, nil)
	fmt.Println(extensionCount, " Extensions found\n")
	extensions := make([]vk.ExtensionProperties, extensionCount)
	vk.EnumerateInstanceExtensionProperties("", &extensionCount, extensions)
	for idx, e := range extensions {
		e.Deref()
		var name string = vk.ToString(e.ExtensionName[:])
		fmt.Printf("%v\t: %v\n", idx+1, name)
	}
}

func xPrintInstanceLayerProperties() {
	var layerCount uint32
	vk.EnumerateInstanceLayerProperties(&layerCount, nil)
	fmt.Println(layerCount, " Layers found\n")
	instanceLayers := make([]vk.LayerProperties, layerCount)
	vk.EnumerateInstanceLayerProperties(&layerCount, instanceLayers)
	for idx, lp := range instanceLayers {
		lp.Deref()
		var name string = vk.ToString(lp.LayerName[:])
		var desc string = vk.ToString(lp.Description[:])
		fmt.Println("Layer ", idx, ": ", name)
		fmt.Println(" Description ", desc)
	}
}

func orPanic(err interface{}) {
	switch v := err.(type) {
	case error:
		if v != nil {
			panic(err)
		}
	case vk.Result:
		if err := vk.Error(v); err != nil {
			panic(err)
		}
	case bool:
		if !v {
			panic("condition failed: != true")
		}
	}
}
