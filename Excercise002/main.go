package main

import (
	"fmt"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

type appObject struct {
	window        *glfw.Window
	instance      vk.Instance
	surface       vk.Surface
	logicaldevice vk.Device
}

// type deviceInfo struct {
// 	gpus     []vk.PhysicalDevice
// 	instance vk.Instance
// 	surface  vk.Surface
// 	device   vk.Device
// }

// Vulkan Device Queue family capabilities Enum
// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/man/html/VkQueueFlagBits.html
type VkQueueFlagBits uint32

const (
	VK_QUEUE_GRAPHICS_BIT VkQueueFlagBits = 1 << iota
	VK_QUEUE_COMPUTE_BIT
	VK_QUEUE_TRANSFER_BIT
	VK_QUEUE_SPARSE_BINDING_BIT
	VK_QUEUE_PROTECTED_BIT
	VK_QUEUE_FLAG_BITS_MAX_ENUM = 0xFFFFFFFF >> 1
)

// Device Queue Capabilities
// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/man/html/VkQueueFlagBits.html

func main() {
	//-------------------------
	// Initialization Stage
	//-------------------------

	var app appObject

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

	//xPrintInstanceLayerProperties()
	//xPrintInstanceExtensionProperties()

	// *** 2 GLFW Window Creation ***//
	// Vulkan WSI extensions avaibale for different platforms
	// https://www.glfw.org/docs/latest/compat.html
	app.window = xCreateWindowGLFW()

	// *** 3 Instance and Application Creation ***//

	app.instance = xCreateInstance()
	fmt.Println(app.instance)

	//xDevicesInfo(instance)
	devices, _ := xGetDevices(app.instance)
	xGetDeviceQueueFamilyProperties(devices[0])

	app.logicaldevice, _ = xCreateLogicalDevice(app.instance)

	// Search for Graphics queue that is capable for supporting Graphics Operations
	xCreateSurface(&app)

	// Cleanup task
	vk.DestroySurface(app.instance, app.surface, nil)
	app.window.Destroy()
	vk.DestroyDevice(app.logicaldevice, nil)
	vk.DestroyInstance(app.instance, nil)
}

func xCreateSurface(app *appObject) {
	surfacePtr, err := app.window.CreateWindowSurface(app.instance, nil)
	if err != nil {
		fmt.Println("Error creating windows surface ", err)
		app.surface = vk.NullSurface
	}
	app.surface = vk.SurfaceFromPointer(surfacePtr)
}

func xCreateWindowGLFW() *glfw.Window {
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err := glfw.CreateWindow(800, 600, "My Game Engine", nil, nil)
	if err != nil {
		fmt.Println("Failed to create window with error ", err)
		return nil
	}
	return window
}

func xCreateLogicalDevice(instance vk.Instance) (vk.Device, error) {
	devices, err := xGetDevices(instance)
	if err != nil {
		err = fmt.Errorf("Failed to get list of physical devices with error: %s", err)
		return nil, err
	}
	// https://www.khronos.org/registry/vulkan/specs/1.2-extensions/html/vkspec.html#VkDeviceQueueCreateInfo
	// See output of 'xGetDeviceQueueFamilyProperties()' to see more details
	var deviceQueueCreateInfoSlice []vk.DeviceQueueCreateInfo = []vk.DeviceQueueCreateInfo{
		{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: 0,
			QueueCount:       1,
			PQueuePriorities: []float32{1.0},
			Flags:            0x7FFFFFFF, // This is equivalent of 'VK_DEVICE_QUEUE_CREATE_FLAG_BITS_MAX_ENUM' bit set or in other words, enable everything available
		},
		{
			SType:            vk.StructureTypeDeviceQueueCreateInfo,
			QueueFamilyIndex: 1,
			QueueCount:       1,
			PQueuePriorities: []float32{1.0},
			Flags:            0x7FFFFFFF,
		},
	}
	var deviceExtensions = []string{"VK_KHR_surface\x00"}
	var deviceLayers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var deviceCreateInfo *vk.DeviceCreateInfo = &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(deviceQueueCreateInfoSlice)),
		PQueueCreateInfos:       deviceQueueCreateInfoSlice,
		EnabledLayerCount:       uint32(len(deviceLayers)),
		PpEnabledLayerNames:     deviceLayers,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
	}
	var logicalDevice vk.Device
	err = vk.Error(vk.CreateDevice(devices[0], deviceCreateInfo, nil, &logicalDevice))
	if err != nil {
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return nil, err
	}
	return logicalDevice, nil
}

func xGetDeviceQueueFamilyProperties(device vk.PhysicalDevice) {
	var familyPropertyCount uint32
	//var gpuqueuecapabilities VkQueueFlagBits
	var gpuqueuecapabilities uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &familyPropertyCount, nil)
	fmt.Println(familyPropertyCount, " Queue Family/Families found on GPU")
	var gpuQueueFamilyProperties = make([]vk.QueueFamilyProperties, familyPropertyCount)
	vk.GetPhysicalDeviceQueueFamilyProperties(device, &familyPropertyCount, gpuQueueFamilyProperties)
	fmt.Println("Following GPU Queue Capabilities were found ....")
	for idx, q := range gpuQueueFamilyProperties {
		q.Deref()
		//f := vk.ToString(q.QueueFlags)
		//fmt.Printf("%v\t: %v\n", idx+1, f)
		fmt.Printf("Flag(s) value %v is(are) set for Queue %v\n", q.QueueFlags, idx)
		gpuqueuecapabilities = uint32(q.QueueFlags)
		if gpuqueuecapabilities&uint32(VK_QUEUE_GRAPHICS_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_GRAPHICS_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_COMPUTE_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_COMPUTE_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_TRANSFER_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_TRANSFER_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_SPARSE_BINDING_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_SPARSE_BINDING_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_PROTECTED_BIT) != 0 {
			fmt.Println("\tVK_QUEUE_PROTECTED_BIT ")
		}
		if gpuqueuecapabilities&uint32(VK_QUEUE_FLAG_BITS_MAX_ENUM) != 0 {
			fmt.Println("\tVK_QUEUE_FLAG_BITS_MAX_ENUM ")
		}
	}

}

func xGetDevices(instance vk.Instance) ([]vk.PhysicalDevice, error) {

	var deviceCount uint32
	err := vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, nil))
	if err != nil {
		err = fmt.Errorf("Failed to get list of physical devices with error: %s", err)
		return nil, err
	}
	fmt.Println(deviceCount, " Physical Device(s) found")
	var devices = make([]vk.PhysicalDevice, deviceCount)
	err = vk.Error(vk.EnumeratePhysicalDevices(instance, &deviceCount, devices))
	if err != nil {
		err = fmt.Errorf("Failed to get properties of physical devices with error: %s", err)
		return nil, err
	}
	return devices, nil
}

func xCreateInstance() vk.Instance {
	var appInfo *vk.ApplicationInfo = &vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.MakeVersion(1, 0, 0), // Throws 'vulkan error: incompatible driver' error with incorrect version number
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	var instance vk.Instance
	var layers = []string{"VK_LAYER_KHRONOS_validation\x00"}
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType: vk.StructureTypeInstanceCreateInfo,
		PNext: nil,
		//Flags:                   nil,
		PApplicationInfo:        appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	orPanic(vk.CreateInstance(&instanceInfo, nil, &instance))
	//InitInstance obtains instance PFNs for Vulkan API functions, this is necessary on OS X using MoltenVK, but for the other platforms it's an option. Not implemented for Android.
	vk.InitInstance(instance)
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
