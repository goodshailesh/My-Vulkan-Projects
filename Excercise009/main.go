package main

import (
	"fmt"

	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
)

func main() {
	glfw.Init()
	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())
	vk.Init()

	//1. Instance
	//	1. Create Application Struct
	//	2. Create Instance
	//	3. Create Window
	//	4. Create Surface
	var instance vk.Instance
	var appInfo vk.ApplicationInfo = vk.ApplicationInfo{
		SType:              vk.StructureTypeApplicationInfo,
		PNext:              nil,
		PApplicationName:   "myVulkan Application\x00",
		ApiVersion:         vk.ApiVersion10,
		ApplicationVersion: vk.MakeVersion(1, 0, 0),
		PEngineName:        "My Game Engine\x00",
		EngineVersion:      vk.MakeVersion(0, 1, 0),
	}
	var layers = []string{"VK_LAYER_NV_optimus\x00"}
	var extensions = []string{"VK_KHR_surface\x00", "VK_KHR_win32_surface\x00"}
	var instanceInfo = vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PNext:                   nil,
		Flags:                   0,
		PApplicationInfo:        &appInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(extensions)),
		PpEnabledExtensionNames: extensions,
	}
	result := vk.CreateInstance(&instanceInfo, nil, &instance)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error creating instance: %v", result))
	}
	var window *glfw.Window
	var err error
	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	window, err = glfw.CreateWindow(640, 480, "Vulkan Info", nil, nil)
	if err != nil {
		fmt.Println("Failed to create window with error :", err)
	}
	var surface vk.Surface
	pSurface, glfwError := window.CreateWindowSurface(instance, nil)
	if glfwError != nil {
		fmt.Println("Failed to create window surface with error :", glfwError)
	}
	surface = vk.SurfaceFromPointer(pSurface)
	fmt.Println(surface)
	//2. Logical Device
	//	1. Get All Physical Devices GPU
	//  2. List Physical Devices GPU properties
	//  3. Get MemoryProprties of Selected Physical Device
	//	4. Get Device Extension Properties
	//	5. Create Logical Device
	var deviceCount uint32
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	fmt.Printf("\tFound total %v Physical Device(s).....\n", deviceCount)
	var physicalDevices = make([]vk.PhysicalDevice, deviceCount)
	result = vk.EnumeratePhysicalDevices(instance, &deviceCount, physicalDevices)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var physicalDevice = physicalDevices[0]
	var physicalDeviceMemoryProperties vk.PhysicalDeviceMemoryProperties
	vk.GetPhysicalDeviceMemoryProperties(physicalDevice, &physicalDeviceMemoryProperties)
	deviceQueueCreateInfo := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: 0,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	var deviceProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(physicalDevice, &deviceProperties)
	deviceProperties.Deref()
	fmt.Println("Driver Details...............")
	fmt.Println("\t* Driver:\t", deviceProperties.DriverVersion)
	fmt.Println("\t* Type:\t\t", deviceProperties.DeviceType)
	fmt.Printf("\t* Version: \t%v.%v.%v\n", (deviceProperties.ApiVersion>>22)&0x3FF, (deviceProperties.ApiVersion>>22)&0x3FF, deviceProperties.ApiVersion&0xFFF)
	fmt.Println("\t* Name:\t", string(deviceProperties.DeviceName[:]))
	var deviceExtensionCount uint32
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, nil)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device count: %v", result))
		panic(result)
	}
	var deviceExtensions = make([]vk.ExtensionProperties, deviceExtensionCount)
	var deviceExtensionNames = make([]string, deviceExtensionCount)
	for idx, extension := range deviceExtensions {
		extension.Deref()
		deviceExtensionNames[idx] = fmt.Sprintf("%v\x00", extension.ExtensionName[:])
	}
	result = vk.EnumerateDeviceExtensionProperties(physicalDevice, "", &deviceExtensionCount, deviceExtensions)
	var deviceFeatures = make([]vk.PhysicalDeviceFeatures, 1)
	deviceFeatures[0].ShaderClipDistance = vk.True
	var deviceCreateInfo vk.DeviceCreateInfo = vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(deviceQueueCreateInfo)),
		PQueueCreateInfos:       deviceQueueCreateInfo,
		EnabledLayerCount:       uint32(len(layers)),
		PpEnabledLayerNames:     layers,
		EnabledExtensionCount:   uint32(len(deviceExtensionNames)),
		PpEnabledExtensionNames: deviceExtensionNames,
		PEnabledFeatures:        deviceFeatures,
	}
	var logicalDevice vk.Device
	result = vk.CreateDevice(physicalDevice, &deviceCreateInfo, nil, &logicalDevice)

	//3. Create SwapChain
	//	1. Get surface capabilities
	//  2. Create Swapchain
	//  3. Get MemoryProprties of Selected Physical Device
	//	4. Get Device Extension Properties
	//	5. Create Logical Device
	var physicalDeviceSurfaceCapabilities vk.SurfaceCapabilities
	result = vk.GetPhysicalDeviceSurfaceCapabilities(physicalDevice, surface, &physicalDeviceSurfaceCapabilities)
	if result != vk.Success {
		fmt.Println(fmt.Errorf("Error getting physical device surface capabilities: %v", result))
		panic(result)
	}
	var surfaceResuloution vk.Extent2D
	surfaceResuloution = physicalDeviceSurfaceCapabilities.CurrentExtent
	// width := surfaceResuloution.Width
	// height := surfaceResuloution.Height
	var swapChainInfo = vk.SwapchainCreateInfo{
		SType:            vk.StructureTypeSwapchainCreateInfo,
		Surface:          surface,
		MinImageCount:    2,
		ImageFormat:      vk.FormatB8g8r8a8Unorm,
		ImageColorSpace:  vk.ColorspaceSrgbNonlinear,
		ImageExtent:      surfaceResuloution,
		ImageArrayLayers: 1,
		ImageUsage:       vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit),
		ImageSharingMode: vk.SharingModeExclusive,
		PreTransform:     vk.SurfaceTransformIdentityBit,
		CompositeAlpha:   vk.CompositeAlphaOpaqueBit,
		PresentMode:      vk.PresentModeMailbox,
		Clipped:          vk.True,
		OldSwapchain:     nil,
	}
	var swapChain vk.Swapchain
	result = vk.CreateSwapchain(logicalDevice, &swapChainInfo, nil, &swapChain)
	fmt.Println(swapChain)

	//Cleanup
	vk.DestroySwapchain(logicalDevice, swapChain, nil)
	vk.DestroySurface(instance, surface, nil)
	vk.DestroyInstance(instance, nil)
	window.Destroy()
}
